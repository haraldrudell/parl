/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import "sync/atomic"

const (
	CyclicAwaitableClosed = true
)

// CyclicAwaitable is an awaitable that can be re-initialized
//   - one-to-many, happens-before
//   - the synchronization mechanic is closing channel, allowing threads to await
//     multiple events
//   - status can be inspected in a thread-safe manner: isClosed, isAboutToClose
//     allows for race-free consumers
//   - Close is idempotent, panic-free
//   - if atomic.Pointer[Awaitable] is used for retrieval, a cyclic semaphore is achieved
type CyclicAwaitable struct {
	p atomic.Pointer[Awaitable]
}

// NewCyclicAwaitable returns an awaitable that can be re-initialized
//   - Init must be invoked prior to use
func NewCyclicAwaitable() (awaitable *CyclicAwaitable) {
	return &CyclicAwaitable{}
}

// Init sets the initial state of the awaitable
//   - default is not triggered
//   - if argument [task.CyclicAwaitableClosed], initial state
//     is triggered
func (a *CyclicAwaitable) Init(initiallyClosed ...bool) (a2 *CyclicAwaitable) {
	a2 = a
	var shouldBeClosed = len(initiallyClosed) > 0 && initiallyClosed[0]
	var awaitable = NewAwaitable()
	if shouldBeClosed {
		awaitable.Close()
	}
	a.p.Store(awaitable)
	return
}

// Ch returns an awaitable channel. Thread-safe
func (a *CyclicAwaitable) Ch() (ch AwaitableCh) {
	return a.p.Load().Ch()
}

// isClosed inspects whether the awaitable has been triggered
//   - isClosed indicates that the channel is closed
//   - isAboutToClose indicates that Close has been invoked,
//     but that channel close may still be in progress
//   - the two values are requried to attain race-free consumers
//   - if isClosed is true, isAboutToClose is also true
//   - Thread-safe
func (a *CyclicAwaitable) IsClosed() (isClosed, isAboutToClose bool) {
	return a.p.Load().IsClosed()
}

// Close triggers awaitable by closing the channel
//   - upon return, the channel is guarantee to be closed
//   - idempotent, panic-free, thread-safe
func (a *CyclicAwaitable) Close() (wasClosed bool) {
	return a.p.Load().Close()
}

// Open rearms the awaitable for another cycle
//   - upon return, the channel is guarantee to be open
//   - idempotent, panic-free, thread-safe
func (a *CyclicAwaitable) Open() (didOpen bool) {
	var openp *Awaitable
	for {
		var ap = a.p.Load()
		var isClosed, _ = ap.IsClosed()
		if !isClosed {
			return // was open return
		}
		if openp == nil {
			openp = NewAwaitable()
		}
		if didOpen = a.p.CompareAndSwap(ap, openp); didOpen {
			return // did open the channel return
		}
	}
}
