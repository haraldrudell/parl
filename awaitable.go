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
// and await any number of events in parallel
//   - [Awaitable.Ch] returns an awaitable channel closing on trig of awaitable.
//     The initial channel state is open
//   - [Awaitable.Close] triggers the awaitable, ie. closes the channel.
//     Upon return, the channel is guaranteed to be closed
//   - — with optional [EvCon] argument, Close is eventually consistent, ie.
//     Close may return prior to channel actually closed
//     for higher performance
//   - [Awaitable.IsClosed] returns whether the awaitable is triggered, ie. if the channel is closed
//   - initialization-free, one-to-many wait mechanic, synchronizes-before, observable
//   - use of channel as mechanic allows consumers to await multiple events
//   - Awaitable costs a lazy channel and pointer allocation
//   - note: [parl.CyclicAwaitable] is re-armable, cyclic version
//   - —
//   - alternative low-blocking inter-thread mechanics are [sync.WaitGroup] and [sync.RWMutex]
//   - — neither is observable and the consumer cannot await other events in parallel
//   - — RWMutex cyclic use has inversion of control issues
//   - — WaitGroup lacks control over waiting threads requiring cyclic use to
//     employ a re-created pointer and value
//   - — both are less performant for the managing thread
type Awaitable struct {
	// ch is lazy atomic-shielded-lock creation of channel
	ch AtomicLockArg[chan struct{}, Awaitable]
	// wg is lazy atomic-shielded-lock creation of wait-group
	//	- wg is separated from ch, reducing contention
	wg AtomicLock[sync.WaitGroup]
	// closeWinner selects the thread to close the channel
	//	- non-zero when close is in progress or completed
	closeWinner atomic.Uint32
	// isClosed is close-complete flag
	//	- atomic-performance shield around the slower channel
	isClosed atomic.Bool
}

// Ch returns an awaitable channel. Thread-safe
//   - ch: non-nil channel
//   - lazy-initialized atomic-shielded-lock
func (a *Awaitable) Ch() (ch AwaitableCh) { return *a.ch.Get(makeCh, a) }

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

	// obtain the channel to check its status
	//	- the channel itself must be checked because
	//		previous invokers of Ch may hold the channel value
	var ch = *a.ch.Get(makeCh, a)

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
//     is atcually closed for higher performance. The close operation is
//     guaranteed to complete in the near future
//   - didClose true: this thread executed close
//   - idempotent, deferrable, panic-free, thread-safe
func (a *Awaitable) Close(eventuallyConsistent ...EventuallyConsistent) (didClose bool) {

	// already closed case
	//	- reading atomic is 0.4955 ns
	if a.isClosed.Load() {
		return // already closed return: didClose false
	}

	// pick very first thread as winner
	//	- Add is faster that CAS
	//	- winner thread is 8.9 ns (read 0.4955 ns + successful CAS 8.419 ns)
	//	- losing thread is 21.5 ns (read 0.4955 ns + failing CAS 21 ns)
	//	- subsequent thread is 0.4955 ns
	var isWinner = a.closeWinner.Load() == 0 && // if already incremented, this thread is not winner
		a.closeWinner.Add(1) == 1 // the first thread to increment is the write is winner

	if !isWinner {

		// eventually consistent case does not wait
		//	- this makes eventually consistent Close a blazing 8.655 ns parallel!
		if len(eventuallyConsistent) > 0 && bool(eventuallyConsistent[0]) {
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
	// single-thread: ≈2 ns
	//	- unshielded parallel contention makes channel read an extremely slow 916 ns
	//	- shielded parallel: 66% is spent in channel read
	if ch := *a.ch.Get(makeCh, a); ch != fakeCh {
		close(ch)
	}
	// close completed

	// on close complete, store atomic performance flag
	a.isClosed.Store(true)
	// release any waiting loser threads
	var wg = a.wg.GetFunc(makeWaitGroup)
	wg.Done()

	didClose = true
	return // didClose return: didClose true
}

// lazyCh creates the channel
//   - *chp is zero-value
//   - invoked maximum once per Awaitable
//   - a.ch.Get delegates to lazyCh
func (a *Awaitable) lazyCh(chp *chan struct{}) {

	// if close in progress, use closed fake static channel
	//	- saves a channel make and channel close
	if a.closeWinner.Load() != 0 {
		*chp = fakeCh
		return
	}
	// create new channel
	//	- invoker is Ch or isClosed while close in progress
	*chp = make(chan struct{})
}

// fakeCh is a static closed channel
var fakeCh = func() (ch chan struct{}) {
	ch = make(chan struct{})
	close(ch)
	return
}()

// makeCh is channel creator function. Thread-safe
//   - Awaitable.ch.Get delegates to makeCh
//   - *chp: zero-value
//   - invoked maximum once per Awaitable
func makeCh(chp *chan struct{}, a *Awaitable) {
	NilPanic("makeCh argument", a)
	a.lazyCh(chp)
}

// makeWaitGroup is wait-group creator function. Thread-safe
//   - Awaitable.wg.GetFunc delegates to makeWaitGroup
//   - *wgp: zero-value
//   - invoked maximum once per Awaitable
func makeWaitGroup(wgp *sync.WaitGroup) { wgp.Add(1) }
