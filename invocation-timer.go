/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/haraldrudell/parl/ptime"
)

const (
	// timer thread checks for hung invocations every 10 seconds
	//	- minimum value, can be set to longer
	defaultTimer = 10 * time.Second
)

// CBFunc is a thread-safe function invoked on
//   - reason == ITParallelism: parallelism exceeding parallelismWarningPoint
//   - reason == ITLatency: latency of an ongoing or just ended invocation
//     exceeds latencyWarningPoint
type CBFunc func(reason CBReason, maxParallelism uint64, maxLatency time.Duration, threadID ThreadID)

// InvocationTimer monitors funtion invocations for parallelism and latency
//   - callback is invoked on exceeding thresholds and reaching a new max
//   - runs one thread per instance while an invocation is active
type InvocationTimer[T any] struct {
	callback    CBFunc
	endCb       func(T)
	timerPeriod time.Duration
	goGen       GoGen

	pointerLock sync.Mutex
	// pointer to linked list of Invocations
	//	- head is read to find oldest current invocation
	//	- head is written to insert, update or delete oldest invocation
	//	- tail is used to insert, update or delete newest invocation
	head atomic.Pointer[Invocation[T]] // written behind pointerlock
	tail *Invocation[T]                // behind pointerLock

	invos       AtomicCounter
	latency     AtomicMax[time.Duration]
	parallelism AtomicMax[uint64]

	threadLock sync.Mutex
	subGo      SubGo // behind threadLock
}

// NewInvocationTimer returns an object alerting of max latency and parallelism
//   - Do is used for new invocations
func NewInvocationTimer[T any](
	callback CBFunc, endCb func(T),
	latencyWarningPoint time.Duration, parallelismWarningPoint uint64,
	timerPeriod time.Duration,
	goGen GoGen,
	fieldp ...*InvocationTimer[T],
) (invokeTimer *InvocationTimer[T]) {
	if callback == nil {
		panic(NilError("callback"))
	}

	if len(fieldp) > 0 {
		invokeTimer = fieldp[0]
	}
	if invokeTimer == nil {
		invokeTimer = &InvocationTimer[T]{}
	}

	if timerPeriod < defaultTimer {
		timerPeriod = defaultTimer
	}
	*invokeTimer = InvocationTimer[T]{
		callback:    callback,
		endCb:       endCb,
		timerPeriod: timerPeriod,
		goGen:       goGen,
	}
	NewAtomicMaxp(&invokeTimer.latency, latencyWarningPoint)
	NewAtomicMaxp(&invokeTimer.parallelism, parallelismWarningPoint)
	return
}

// Oldest returns the oldest invocation
//   - threadID is ID of oldest active thread, if any
//   - age is longest ever invocation
//   - if no invocation is active, age is 0, threadID invalid
func (i *InvocationTimer[T]) Oldest() (age time.Duration, threadID ThreadID) {

	// get any active invocation
	var invocation = i.head.Load()
	if invocation == nil {
		return // no active invocation return
	}
	threadID = invocation.ThreadID

	// get age of oldest active invocation
	age = invocation.Age()
	if age2, _ := i.latency.Max(); age2 > age {
		age = age2
	}

	return
}

// Invocation registers a new invocation with callbacks for parallelism and latency
//   - caller invokes deferFunc at end of invocation
//
// Usage:
//
//	func someFunc() {
//	  defer invocationTimer.Invocation()()
func (i *InvocationTimer[T]) Invocation(value T) (deferFunc func()) {
	var invocation = NewInvocation(i.invocationEnd, value) // one allocation
	i.insert(invocation)
	i.ensureTimer()
	var invos = i.invos.Value()
	if i.parallelism.Value(invos) {
		var max, _ = i.latency.Max()
		// callback for high parallelism warning
		i.callback(ITParallelism, invos, max, invocation.ThreadID)
	}
	return invocation.DeferFunc // one allocation
}

// invocationEnd is invoked by the Invocation instance’s deferred function
func (i *InvocationTimer[T]) invocationEnd(invocation *Invocation[T], duration time.Duration) {
	i.remove(invocation)
	i.maybeCancelTimer()
	if i.latency.Value(duration) {
		var max, _ = i.parallelism.Max()
		// callback for slowness of completed task
		i.callback(ITLatency, max, duration, invocation.ThreadID)
	}
	if cb := i.endCb; cb != nil {
		cb(invocation.Value)
	}
}

func (i *InvocationTimer[T]) insert(invocation *Invocation[T]) {
	i.pointerLock.Lock()
	defer i.pointerLock.Unlock()

	// link in at tail
	var tail = i.tail
	if tail != nil {
		invocation.Prev.Store(tail)
		tail.Next.Store(invocation)
	}
	i.tail = invocation

	// if first item, update head
	if tail == nil {
		i.head.Store(invocation)
	}
}

func (i *InvocationTimer[T]) remove(invocation *Invocation[T]) {
	i.pointerLock.Lock()
	defer i.pointerLock.Unlock()

	var prev = invocation.Prev.Load()
	var next = invocation.Next.Load()

	// unlink at previous item
	if prev == nil {
		i.head.Store(next)
	} else {
		prev.Next.Store(next)
	}

	// unlink at next item
	if next == nil {
		i.tail = prev
	} else {
		next.Prev.Store(prev)
	}
}

// ensureTimer ensures that a time is eventually running if it should be
func (i *InvocationTimer[T]) ensureTimer() {
	// if this was not the first invocation from idle,
	// a thread does not have to be launched
	if i.invos.Inc() != 1 {
		return // this was not the initial invocation
	}

	i.threadLock.Lock()
	defer i.threadLock.Unlock()

	if i.invos.Value() == 0 {
		return // other threads decremented value to zero
	} else if i.subGo != nil {
		return // some other thread already launched the timer thread
	}

	// order thread launch
	var subGo = i.goGen.SubGo()
	i.subGo = subGo
	go i.hungInvocationCheckThread(ptime.NewOnTicker(i.timerPeriod, time.Local), subGo.Go())
}

// maybeCancelTimer ensures that any timer thread ordered to launch will exit
func (i *InvocationTimer[T]) maybeCancelTimer() {
	// if the number iof invocations does not go to zero,
	// a thread does not need to be stopped
	if i.invos.Dec() != 0 {
		return // more invocations are active
	}

	i.threadLock.Lock()
	defer i.threadLock.Unlock()

	if i.invos.Value() != 0 {
		return // another thread launched invocations
	}

	// cancel timer thread
	var subGo = i.subGo
	if subGo == nil {
		return // another thread already shut down the timer thread
	}
	i.subGo = nil
	subGo.Cancel()
}

// hungInvocationCheckThread looks for invocations that do not return
func (i *InvocationTimer[T]) hungInvocationCheckThread(ticker *ptime.OnTicker, g Go) {
	var err error
	defer g.Register().Done(&err)
	defer RecoverErr(func() DA { return A() }, &err)

	C := ticker.C
	done := g.Context().Done()
	for {
		select {
		case <-done:
			return
		case <-C:
		}

		// get oldest active invocation
		var oldestInvocation = i.head.Load()
		if oldestInvocation == nil {
			continue // noop: no invocation
		}

		// check if it is oldest yet
		var age = oldestInvocation.Age()
		if !i.latency.Value(age) {
			continue // not oldest yet
		}

		// invoke callback
		var max, _ = i.parallelism.Max()
		// callback for high latency of task in progress
		i.callback(ITLatency, max, age, oldestInvocation.ThreadID)
	}
}
