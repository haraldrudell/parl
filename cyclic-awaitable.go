/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import "sync/atomic"

// CyclicAwaitable is an awaitable that can be re-initialized
//   - one-to-many, happens-before
//   - the synchronization mechanic is closing channel, allowing consumers to await
//     multiple events
//   - [CyclicAwaitable.IsClosed] provides thread-safe observability
//   - [CyclicAwaitable.Close] is idempotent, thread-safe and deferrable
//   - Open means event is pending, Close means event has triggered
//   - [CyclicAwaitable.Open] arms the awaitable returning a channel guaranteed to be
//     open at timeof invocation
//   - —
//   - because Awaitable is renewed, access is via atomic Pointer
//   - Pointer to struct allows for atomic update of IsClosed and Open
//
// Usage:
//
//	valueWaiter *CyclicAwaitable
//	…
//	func (v *V) getOrWaitForValue(value T) {
//	  var hasValue bool
//	  // check if value is already present
//	  if value, hasValue = v.hasValueFromThread(); hasValue {
//	    return
//	  }
//	  // arm cyclable
//	  v.valueWaiter.Open()
//	  // collect any value arriving prior to arming cyclable
//	  if value, hasValue = v.hasValueFromThread(); hasValue {
//	    return
//	  }
//	  <- v.valueWaiter.Ch()
//	  value, _ = v.hasValueFromThread()
//	…
//	func (v *V) threadStoresValue(value T) {
//	  v.store(value)
//	  v.valueWaiter.Close()
type CyclicAwaitable struct{ awp atomic.Pointer[Awaitable] }

// Ch returns an awaitable channel. Thread-safe
func (a *CyclicAwaitable) Ch() (ch AwaitableCh) { return a.aw().Ch() }

// isClosed inspects whether the awaitable has been triggered
//   - isClosed indicates that the channel is closed
//   - Thread-safe
func (a *CyclicAwaitable) IsClosed() (isClosed bool) { return a.aw().IsClosed() }

// Close triggers awaitable by closing the channel
//   - upon return, the channel is guaranteed to be closed
//   - idempotent, thread-safe
func (a *CyclicAwaitable) Close() (didClose bool) { return a.aw().Close() }

// Open rearms the awaitable for another cycle
//   - ch is guaranteed to have been open at time of invocation
//   - didOpen is true if the channel was encountered closed
//   - idempotent, thread-safe
func (a *CyclicAwaitable) Open() (didOpen bool, ch AwaitableCh) {
	var openAwaitable Awaitable
	for {
		var awaitable = a.awp.Load()
		if awaitable != nil && !awaitable.IsClosed() {
			ch = awaitable.Ch()
			return // was open return
		}
		if didOpen = a.awp.CompareAndSwap(awaitable, &openAwaitable); didOpen {
			ch = openAwaitable.Ch()
			return // did open the channel return
		}
	}
}

// aw returns the active awaitable using atomic mechanic
func (a *CyclicAwaitable) aw() (aw *Awaitable) {
	if a == nil {
		panic(NilError("CyclicAwaitable pointer"))
	}
	var newAwaitable Awaitable
	for {
		if aw = a.awp.Load(); aw != nil {
			return // existing awaitable return
		} else if a.awp.CompareAndSwap(nil, &newAwaitable) {
			aw = &newAwaitable
			return // wrote new awaitable return
		}
	}
}
