/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import (
	"sync"
	"sync/atomic"
)

// Awaitable is a semaphore allowing any number of threads to observe
// and await any number of events in parallel: wait-free-locked
//   - performance: the only allocation is a channel on:
//   - — Ch prior to Close
//   - — rare IsClose events that will no happen
//   - — other than that one-to-low-ns everything
//   - [Awaitable.Ch] returns an awaitable channel closing on trig of awaitable.
//     The initial channel state is open
//   - [Awaitable.Close] triggers the awaitable, ie. closes the channel.
//     Upon return, the channel is guaranteed to be closed
//   - — obsolete, still present: with optional [EvCon] argument,
//     Close is eventually consistent
//   - [Awaitable.IsClosed] returns whether the awaitable is triggered, ie. if the channel is closed
//   - initialization-free, one-to-many wait mechanic, synchronizes-before, inspectable
//   - use of channel as mechanic allows consumers to await multiple events: wait-free-locked
//   - Awaitable costs lazy channel allocation
//   - note: [parl.CyclicAwaitable] is re-armable, cyclic version
//   - —
//   - alternative low-blocking inter-thread mechanics are [sync.WaitGroup] and [sync.RWMutex]
//   - — neither is inspectable and the consumer cannot await multiple events
//   - — RWMutex cyclic use has inversion of control issues
//   - — WaitGroup requires synchronization of Add/Done and Wait invocations or panic
//   - — both are less performant for the managing thread
type Awaitable struct {
	// isGet is true if a channel-read operation was initiated
	//	- a channel is or may soon be available or created
	isGet atomic.Bool
	// ch is lazily initialized, atomic access-performance
	// singleton channel value
	ch AtomicLockArg[chan struct{}, Awaitable]
	// wg ensures no close invocation returns prior to
	// channel close complete
	//	- wg is lazily initialized, atomic access-performance
	//		singleton wait-group
	//	- wg wait is separate from ch, reducing contention
	wg AtomicLock[sync.WaitGroup]
	// closeWinner selects the thread to close the channel
	//	- non-zero when close is in progress or completed
	//	- zero means Close has not been invoked
	closeWinner atomic.Uint32
	// isClosed is set to true upon close-complete
	//	- atomic-performance shield around the slower channel
	isClosed atomic.Bool
}

// Ch returns an awaitable channel. Thread-safe
//   - ch: non-nil channel
func (a *Awaitable) Ch() (ch AwaitableCh) { return a.getCh() }

// isClosed inspects whether the awaitable has been triggered
//   - isClosed true: channel is closed
//   - Thread-safe
func (a *Awaitable) IsClosed() (isClosed bool) {

	// read whether Close has completed at atomic performance
	//	- reading atomic is 0.4955 ns
	if a.isClosed.Load() {
		isClosed = true
		return // already closed return: isClosed true
	}
	// no channel close has completed

	// read whether Close was initiated at atomic performance
	if a.closeWinner.Load() == 0 {
		// no channel close has begun
		return // no close return: isClosed false
	}
	// a close is in progress

	// IsClosed must get its status by reading the channel
	//	- the channel itself must be checked because
	//		Close progress is uncertain
	var ch = a.getCh()

	// check whether close assigned a fake channel
	if ch == fakeCh {
		isClosed = true
		return // faster fake-check return: isClosed true
	}

	// get exact close state from the channel
	//	- determining channel close is 3.479 ns
	select {
	// if channel sends data, it is because it is closed
	case <-ch:
		isClosed = true
	default:
	}

	return // status from channel return: isClosed valid
}

// Close triggers awaitable by closing the channel
//   - eventuallyConsistent missing: upon return, the channel is
//     guaranteed to be closed
//   - eventuallyConsistent [EvCon]: Close may return before the channel
//     is actually closed for higher performance. The close operation is
//     guaranteed to complete in the near future
//   - didClose true: this thread executed close
//   - —
//   - idempotent, deferrable, panic-free, thread-safe
//   - eventual consistency has lost its performance significance
//     due to extensive use of atomics
//   - — close on awaitable prior to Ch invocation is 1 ns from
//     deferred channel creation and use of fake channel
//   - — close of closed is 1 ns
//   - — eventual consistency is only effective in the unlikely hit of a
//     34 ns window for the case when a real channel is being closed
//   - — designing a benchmark exhibiting any benefit is impractical
//   - — prior to 2025 refactor, eventual consistency was significant
func (a *Awaitable) Close(eventuallyConsistent ...EventuallyConsistent) (didClose bool) {

	// already closed case
	//	- reading atomic is 0.4955 ns
	if a.isClosed.Load() {
		return // already closed return: didClose false
	}

	// pick very first thread as winner
	//	- Add is faster than CAS
	//	- winner thread is 8.9 ns (read 0.4955 ns + successful CAS 8.419 ns)
	//	- losing thread is 21.5 ns (read 0.4955 ns + failing CAS 21 ns)
	//	- subsequent thread is 0.4955 ns
	//	- —
	//	- note that subsequent threads proceed 8–20 ns faster than
	//		initial threads
	var isWinner = a.closeWinner.Load() == 0 && // if already incremented, this thread is not winner
		a.closeWinner.Add(1) == 1 // the first thread to increment obtaining 1 is winner

	if !isWinner {

		// eventually consistent case does not wait
		//	- this makes eventually consistent Close a blazing 8.655 ns parallel!
		if len(eventuallyConsistent) > 0 && eventuallyConsistent[0] == EventuallyConsistency {
			// eventually consistent: another thread is closing the channel
			return // eventually consistent return: didClose false
		}
		// this thread should await channel close, there is little hurry

		// wait for shared waitGroup
		var wg = a.wg.GetFunc(makeWaitGroup)
		wg.Wait()
		return // waited return: didClose false
	}
	// this thread should close the channel

	// execute close
	//	- there is atomic race between this [Awaitable.Close] and
	//		Awaitable.getCh
	//	- getCh is invoked via [Awaitable.Ch] or [Awaitable.IsClosed]

	// getCh:
	//	- at top, sets isGet true: a channel is or may be about to be created
	//	- prior to creating the channel, inspects closeWinner
	//	- non-zero closeWinner: returns fake channel: static, closed value
	//	- zero closeWinner: creates a channel value ~30 ns

	// Close:
	//	- at top, sets closeWinner non-zero: the awaitable is about to close
	//	- if isGet is observed false, no channel is required: done
	//	- — once getCh is invoked, it will return fake channel
	//	- if isGet is observed true, the channel must be obtained
	//	- — if returned channel is fake: done
	//	- — if returned channel is real, it has to be closed ~35 ns

	//	- if the channel wasn’t read, its close can be deferred
	//	- isGet checks closeWinner which is already set
	if a.isGet.Load() {
		// the channel was read
		//	- must close any channel that is not fake
		if ch := *a.ch.Get(makeCh, a); ch != fakeCh {
			// single-thread: ≈2 ns
			//	- unshielded parallel contention makes channel read an extremely slow 916 ns
			//	- shielded parallel: 66% is spent in channel read
			close(ch)
		}
	}
	// close complete or channel creation deferred

	// on close complete, store atomic performance flag
	//	- cannot be written prior to close complete
	a.isClosed.Store(true)

	// release any waiting loser threads
	//	- cannot be done prior to close complete
	var wg = a.wg.GetFunc(makeWaitGroup)
	// [sync.WaitGroup.Done] Pike’s best invention:
	//	- no inversion of control
	//	- lock-free to releasing thread
	wg.Done()

	didClose = true
	return // didClose return: didClose true
}

// getCh creates the channel
//   - sets indicator that channel is about to be created
//   - lazy-initialized atomic-shielded-lock
func (a *Awaitable) getCh() (ch chan struct{}) {

	// parallel atomic: avoid Store
	if !a.isGet.Load() {
		a.isGet.Store(true)
	}
	// isGet is true

	// Get retrieves pointer to singleton channel value,
	// possibly invoking makeCh to initialize it
	//	- makeCh delegates to lazyCh method
	ch = *a.ch.Get(makeCh, a)

	return
}

// lazyCh initializes the Awaitable’s channel value
//   - chp: points to unreferenced zero-value
//   - —
//   - lazyCh is invoked maximum once per Awaitable
//   - inside critical section, thread-safe
//   - a.ch.Get delegates to lazyCh
func (a *Awaitable) lazyCh(chp *chan struct{}) {

	// if close in progress, use closed fake static channel
	//	- saves a channel make and channel close
	if a.closeWinner.Load() != 0 {
		*chp = fakeCh
		return
	}

	// create new channel
	//	- invoker is Ch prior to Close or
	//	- IsClosed while Close in progress
	*chp = make(chan struct{})
}

// fakeCh is a static closed channel
var fakeCh = func() (ch chan struct{}) {
	ch = make(chan struct{})
	close(ch)
	return
}()

// makeCh is channel initializer function provided to
// [AtomicLock.Get]
//   - chp: points to unreferenced channel zero-value
//   - a: argument, ie. the Awaitable object
//   - —
//   - inside critical section, thread-safe
//   - invoked maximum once per Awaitable
func makeCh(chp *chan struct{}, a *Awaitable) {

	// panic if a is nil
	NilPanic("makeCh argument", a)

	a.lazyCh(chp)
}

// makeWaitGroup is wait-group initializer function
//   - wgp: pointer to unreferenced [sync.WaitGroup] zero-value
//   - —
//   - inside critical section, thread-safe
//   - Awaitable.wg.GetFunc delegates to makeWaitGroup
//   - makeWaitGroup is invoked maximum once per Awaitable
func makeWaitGroup(wgp *sync.WaitGroup) { wgp.Add(1) }
