/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import "sync/atomic"

const (
	// as argument to NewCyclicAwaitable, causes the awaitable ot be initially
	// triggered
	CyclicAwaitableClosed bool = true
)

// CyclicAwaitable is an awaitable that can be re-initialized
//   - one-to-many, happens-before
//   - the synchronization mechanic is closing channel, allowing threads to await
//     multiple events
//   - status can be inspected in a thread-safe manner: isClosed, isAboutToClose
//     allows for race-free consumers
//   - Close is idempotent, panic-free
//   - if atomic.Pointer[Awaitable] is used for retrieval, a cyclic semaphore is achieved
type CyclicAwaitable struct{ *atomic.Pointer[Awaitable] }

// NewCyclicAwaitable returns an awaitable that can be re-initialized
//   - if argument [task.CyclicAwaitableClosed] is provided, the initial state
//     of the CyclicAwaitable is triggered
func NewCyclicAwaitable(initiallyClosed ...bool) (awaitable *CyclicAwaitable) {
	c := CyclicAwaitable{Pointer: &atomic.Pointer[Awaitable]{}}
	c.Store(NewAwaitable())
	if len(initiallyClosed) > 0 && initiallyClosed[0] {
		c.Close()
	}
	return &c
}

// Ch returns an awaitable channel. Thread-safe
func (a *CyclicAwaitable) Ch() (ch AwaitableCh) { return a.Pointer.Load().Ch() }

// isClosed inspects whether the awaitable has been triggered
//   - isClosed indicates that the channel is closed
//   - Thread-safe
func (a *CyclicAwaitable) IsClosed() (isClosed bool) { return a.Load().IsClosed() }

// Close triggers awaitable by closing the channel
//   - upon return, the channel is guaranteed to be closed
//   - idempotent, panic-free, thread-safe
func (a *CyclicAwaitable) Close() (didClose bool) { return a.Load().Close() }

// Open rearms the awaitable for another cycle
//   - upon return, the channel is guarantee to be open
//   - idempotent, panic-free, thread-safe
func (a *CyclicAwaitable) Open() (didOpen bool) {
	var openedAwaitable *Awaitable
	for {
		var awaitable = a.Load()
		if !awaitable.IsClosed() {
			return // was open return
		}
		if openedAwaitable == nil {
			openedAwaitable = NewAwaitable()
		}
		if didOpen = a.CompareAndSwap(awaitable, openedAwaitable); didOpen {
			return // did open the channel return
		}
	}
}
