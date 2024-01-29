/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import "sync/atomic"

// Awaitable is a semaphore allowing any number of threads to observe
// and await an event
//   - one-to-many, synchronizes-before, initialization-free
//   - the synchronization mechanic is closing channel, allowing consumers to await
//     multiple events
//   - [Awaitable.IsClosed] provides thread-safe observability
//   - [Awaitable.Close] is idempotent, thread-safe, deferrable, initialization-free and panic-free
//   - [parl.CyclicAwaitable] is re-armable, cyclic version
//   - —
//   - alternative low-blocking inter-thread mechanics are [sync.WaitGroup] and [sync.RWMutex]
//     but those are less performant for the managing thread
type Awaitable struct {
	isClosed atomic.Bool
	// channel managed by atomicCh
	chanp atomic.Pointer[chan struct{}]
}

// Ch returns an awaitable channel. Thread-safe
func (a *Awaitable) Ch() (ch AwaitableCh) { return a.atomicCh() }

// isClosed inspects whether the awaitable has been triggered
//   - on true return, it is guaranteed that the channel has been closed
//   - Thread-safe
func (a *Awaitable) IsClosed() (isClosed bool) {
	select {
	case <-a.atomicCh():
		isClosed = true
	default:
	}
	return
}

// Close triggers awaitable by closing the channel
//   - upon return, the channel is guaranteed to be closed
//   - idempotent, deferrable, panic-free, thread-safe
func (a *Awaitable) Close() (didClose bool) {
	var ch = a.atomicCh()
	if didClose = a.isClosed.CompareAndSwap(false, true); !didClose {
		<-ch   // prevent returning before channel close
		return // already closed return
	}
	close(ch)
	return // didClose return
}

// atomicCh returns a non-nil channel using atomic mechanic
func (a *Awaitable) atomicCh() (ch chan struct{}) {
	var newChan chan struct{}
	for {
		var loadedChanp = a.chanp.Load()
		if loadedChanp != nil {
			if ch = *loadedChanp; ch != nil {
				return // channel from atomic pointer
			}
		}
		if newChan == nil {
			newChan = make(chan struct{})
		}
		if a.chanp.CompareAndSwap(loadedChanp, &newChan) {
			ch = newChan
			return // channel written to atomic pointer
		}
	}
}
