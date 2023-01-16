/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
	"time"

	"github.com/haraldrudell/parl/pmaps"
	"github.com/haraldrudell/parl/pslice"
	"github.com/haraldrudell/parl/ptime"
	"github.com/haraldrudell/parl/set"
)

const (
	ITParallelism CBReason = iota + 1
	ITLatency

	minLatencyWarningPoint = 10 * time.Millisecond
	defaultTimer           = 10 * time.Second
)

type CBReason uint8

type emid uint64

var emidGenerator UniqueIDTypedUint64[emid]

type CBFunc func(reason CBReason, maxParallelism uint64, maxLatency time.Duration, threadID ThreadID)

type InvokeTimer struct {
	callback    CBFunc
	timerPeriod time.Duration
	invoTimes   pmaps.ThreadSafeOrderedMapAny[emid, uint64, *invocationX]
	invos       AtomicCounter
	latency     AtomicMax[time.Duration]
	parallelism AtomicMax[uint64]

	goLock   sync.Mutex
	cancelGo func()

	g0 GoGen
}

func NewInvokeTimer(
	callback CBFunc,
	latencyWarningPoint time.Duration,
	parallelismWarningPoint uint64,
	timerPeriod time.Duration, g0 GoGen) (invokeTimer *InvokeTimer) {
	if timerPeriod < defaultTimer {
		timerPeriod = defaultTimer
	}
	var ix invocationX
	i := InvokeTimer{
		timerPeriod: timerPeriod,
		invoTimes:   *pmaps.NewThreadSafeOrderedMapAny[emid](ix.order),
		g0:          g0,
	}
	i.latency.Value(latencyWarningPoint)
	i.parallelism.Value(parallelismWarningPoint)
	return &i
}

func (em *InvokeTimer) Do(fn func()) {
	emID := emidGenerator.ID()
	invocation := invocationX{
		t0:       time.Now(),
		threadID: goID(),
	}
	defer func() {
		em.invoTimes.Delete(emID)
		em.maybeCancelTimer()
		if duration := time.Since(invocation.t0); em.latency.Value(duration) {
			// callback for slowness of completed task
			em.callback(ITLatency, em.parallelism.Max(), duration, invocation.threadID)
		}
	}()

	em.invoTimes.Put(emID, &invocation)
	em.ensureTimer()
	invos := em.invos.Value()
	if em.parallelism.Value(invos) {
		// callback for high parallelism warning
		em.callback(ITParallelism, invos, em.latency.Max(), invocation.threadID)
	}

	// execute
	fn()
}

func (em *InvokeTimer) Oldest() (age time.Duration, threadID ThreadID) {
	list := em.invoTimes.List(1)
	if len(list) == 0 {
		return
	}
	invocation := list[0]

	age = time.Since(invocation.t0)
	if age2 := em.latency.Max(); age2 > age {
		age = age2
	}
	threadID = invocation.threadID

	return
}

func (em *InvokeTimer) ensureTimer() {
	em.goLock.Lock()
	defer em.goLock.Unlock()

	em.invos.Inc()

	if em.cancelGo != nil {
		return // timer already running return
	}

	// launch timer
	g0 := em.g0.Go()
	em.cancelGo = g0.CancelGo // save the cancel function for the goroutine
	go ptime.OnTimedThread(em.timerLatencyCheck, em.timerPeriod, time.Local, g0)
}

func (em *InvokeTimer) maybeCancelTimer() {
	em.goLock.Lock()
	defer em.goLock.Unlock()

	if em.invos.Dec() != 0 {
		return // more invocations are active
	} else if em.cancelGo == nil {
		return // no timer running return
	}

	// cancel timer
	em.cancelGo()
	em.cancelGo = nil
}

func (em *InvokeTimer) timerLatencyCheck(at time.Time) {

	// get age of oldest request being processed
	var age time.Duration
	var threadID ThreadID
	if list := em.invoTimes.List(1); len(list) == 0 {
		return // no requests being processed
	} else {
		invocationXp := list[0]
		age = time.Since(invocationXp.t0)
		threadID = invocationXp.threadID
	}

	// print if slowest yet
	if em.latency.Value(age) {
		// callback for high latency of task in progress
		em.callback(ITLatency, em.parallelism.Max(), age, threadID)
	}
}

type invocationX struct {
	t0       time.Time
	threadID ThreadID
}

func (ix *invocationX) order(a *invocationX) (result uint64) {
	if a.t0.IsZero() {
		return // zero-time: 0
	}
	result = uint64(a.t0.UnixNano())
	return
}

func (ws CBReason) String() (s string) {
	return cbReasonSet.StringT(ws)
}

var cbReasonSet = set.NewSet(pslice.ConvertSliceToInterface[
	set.SetElement[CBReason],
	set.Element[CBReason],
]([]set.SetElement[CBReason]{
	{ValueV: ITParallelism, Name: "max parallel"},
	{ValueV: ITLatency, Name: "slowest"},
}))
