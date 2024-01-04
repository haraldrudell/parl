/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import "sync/atomic"

// Awaitable is a semaphore allowing any number of threads to observe
// and await an event
//   - one-to-many, happens-before, initialization-free
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
	// channel from atomicCh
	cp atomic.Pointer[chan struct{}]
}

// NewAwaitable returns a one-to-many semaphore
func NewAwaitable() (awaitable *Awaitable) { return &Awaitable{} }

// Ch returns an awaitable channel. Thread-safe
func (a *Awaitable) Ch() (ch AwaitableCh) { return a.atomicCh() }

func (a *Awaitable) atomicCh() (ch chan struct{}) {
	var c2 chan struct{}
	for {
		var cp = a.cp.Load()
		if cp != nil {
			if ch = *cp; ch != nil {
				return // channel from atomic pointer
			}
		}
		if c2 == nil {
			c2 = make(chan struct{})
		}
		if a.cp.CompareAndSwap(cp, &c2) {
			ch = c2
			return // channel written to atomic pointer
		}
	}
}

// isClosed inspects whether the awaitable has been triggered
//   - Thread-safe
func (a *Awaitable) IsClosed() (isClosed bool) {
	var ch = a.atomicCh()
	select {
	case <-ch:
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
