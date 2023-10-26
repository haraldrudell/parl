/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import "sync/atomic"

// Awaitable is a semaphore that allows any number of threads to inspect
// and await an event
//   - one-to-many, happens-before
//   - the synchronization mechanic is closing channel, allowing threads to await
//     multiple events
//   - status can be inspected in a thread-safe manner: isClosed, isAboutToClose
//     allows for race-free consumers
//   - Close is idempotent, panic-free
//   - if atomic.Pointer[Awaitable] is used for retrieval, a cyclic semaphore is achieved
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
//   - isAboutToClose indicates that Close has been invoked,
//     but that channel close may still be in progress
//   - the two values are requried to attain race-free consumers
//   - if isClosed is true, isAboutToClose is also true
//   - Thread-safe
func (a *Awaitable) IsClosed() (isClosed, isAboutToClose bool) {
	select {
	case <-a.ch:
		isClosed = true
	default:
	}
	isAboutToClose = a.isClosed.Load()
	return
}

// Close triggers awaitable by closing the channel
//   - upon return, the channel is guarantee to be closed
//   - idempotent, panic-free, thread-safe
func (a *Awaitable) Close() (wasClosed bool) {
	if wasClosed = !a.isClosed.CompareAndSwap(false, true); wasClosed {
		<-a.ch // wait to make certain the channel is closed
		return // already closed return
	}
	close(a.ch)
	return // didClose return
}
