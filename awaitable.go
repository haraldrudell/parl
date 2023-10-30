/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import "sync/atomic"

// Awaitable is a semaphore that allows any number of threads to inspect
// and await an event
//   - one-to-many, happens-before
//   - the synchronization mechanic is closing channel, allowing consumers to await
//     multiple events
//   - IsClosed provides thread-safe observability
//   - Close is idempotent and panic-free
//   - [parl.CyclicAwaitable] is re-armable, cyclic version
type Awaitable struct {
	isClosed atomic.Bool
	ch       chan struct{}
}

// NewAwaitable returns a one-to-many sempahore
func NewAwaitable() (awaitable *Awaitable) {
	return &Awaitable{ch: make(chan struct{})}
}

// Ch returns an awaitable channel. Thread-safe
func (a *Awaitable) Ch() (ch AwaitableCh) {
	return a.ch
}

// isClosed inspects whether the awaitable has been triggered
//   - isClosed indicates that the channel is closed
//   - Thread-safe
func (a *Awaitable) IsClosed() (isClosed bool) {
	select {
	case <-a.ch:
		isClosed = true
	default:
	}
	return
}

// Close triggers awaitable by closing the channel
//   - upon return, the channel is guaranteed to be closed
//   - idempotent, panic-free, thread-safe
func (a *Awaitable) Close() (didClose bool) {
	if didClose = a.isClosed.CompareAndSwap(false, true); !didClose {
		<-a.ch // wait to make certain the channel is closed
		return // already closed return
	}
	close(a.ch)
	return // didClose return
}
