/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package cyclebreaker

import "sync/atomic"

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
	// closeWinner selects the thread to close the channel
	closeWinner atomic.Bool
	// channel managed by atomicCh
	//	- lazy initialization
	chanp atomic.Pointer[chan struct{}]
	// isClosed indicates whether the channel is closed at atomic performance
	//	- set to true after channel close complete
	//	- shields channel close detection
	isClosed atomic.Bool
}

// Ch returns an awaitable channel. Thread-safe
func (a *Awaitable) Ch() (ch AwaitableCh) { return a.atomicCh() }

// isClosed inspects whether the awaitable has been triggered
//   - on true return, it is guaranteed that the channel has been closed
//   - Thread-safe
func (a *Awaitable) IsClosed() (isClosed bool) {

	// read close state with atomic performance
	//	- reading atomic is 0.4955 ns
	if isClosed = a.isClosed.Load(); isClosed {
		return
	}

	// get exact close state from the channel
	//	- determining channel close is 3.479 ns
	select {
	case <-a.atomicCh():
		isClosed = true
	default:
	}

	return
}

// [Awaitable.Close] argument to Close meaning eventually consistency
//   - may return before the channel is actually closed
const EvCon = true

// Close triggers awaitable by closing the channel
//   - upon return, the channel is guaranteed to be closed
//   - eventuallyConsistent [EvCon]: may return before the channel is atcually closed
//     for higher performance
//   - idempotent, deferrable, panic-free, thread-safe
func (a *Awaitable) Close(eventuallyConsistent ...bool) (didClose bool) {

	// select close winner
	//	- CAS fail is 21.195 ns, CAS success if 8.419 ns
	//	- atomic read: 0.4955 ns
	if didClose = !( //
	// this thread does not close if:
	//	- a winner was already selected, atomic Load performance or
	a.closeWinner.Load() ||
		// this thread is not the winner
		!a.closeWinner.CompareAndSwap(false, true)); //
	!didClose {

		// eventually consistent case does not wait
		//	- this makes eventually consistent Close a blazing 8.655 ns parallel!
		if len(eventuallyConsistent) > 0 && eventuallyConsistent[0] {
			return // eventually consistent: another thread is closing the channel
		}

		// prevent returning before channel close
		//	- closing thread successful CAS and channel close is 17 ns
		//	- losing thread failing CAS is 21 ns
		//	- the channel is likely already closed
		if a.isClosed.Load() {
			return
		}

		// single-thread: ≈2 ns
		//	- unshielded parallel contention makes channel read an extremely slow 916 ns
		//	- shielded parallel: 66% is spent in channel read
		<-a.atomicCh()
		return // close completed by other thread return
	}
	// only the winner thread arrives here

	// channel close
	//	- ≈9 ns
	close(a.atomicCh())
	// on close complete, store atomic performance flag
	a.isClosed.Store(true)

	return // didClose return
}

// atomicCh returns a non-nil channel using atomic mechanic
func (a *Awaitable) atomicCh() (ch chan struct{}) {

	// get channel previously created by another thread
	//	- 1-pointer Load 0.5167 ns
	if cp := a.chanp.Load(); cp != nil {
		return *cp // channel from atomic pointer
	}

	// attempt to create the authoritative channel
	//	- make of channel is 21.10 ns, 31.13 ns parallel
	//	- CAS fail is 21.195 ns, CAS success if 8.419 ns
	if ch2 := make(chan struct{}); a.chanp.CompareAndSwap(nil, &ch2) {
		return ch2 // channel written to atomic pointer
	}

	// get channel created by other thread
	return *a.chanp.Load()
}
