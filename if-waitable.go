/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "fmt"

// WaitedOn are all goroutine methods provided by [WaitGroupCh]
type WaitedOn interface {
	// Add()
	SyncAdd
	// Done()
	SyncDone
	// DoneBool decrements the WaitGroup counter by one
	//   - isExit true: the counter reached zero
	//   - DoneBool when counter already zero is panic
	DoneBool() (isExit bool)
}

// WaitingFor are all consumer methods provided by [WaitGroupCh]
type WaitingFor interface {
	// Add()
	SyncAdd
	// Ch returns a channel that closes once the counter reaches zero
	Ch() (awaitableCh AwaitableCh)
	// Count returns the current number of remaining threads
	Count() (currentCount int)
	// Counts returns the current state optionally adjusting the counter
	//   - delta is optional counter adjustment
	//   - — delta negative and larger than current count is panic
	//   - — delta present after the counter reduced to zero is panic
	//   - currentCount: remaining count after any delta applied
	//   - totalAdds: cumulative positive adds over WaitGroup lifetime after amny delta applied
	Counts(delta ...int) (currentCount, totalAdds int)
	// IsZero returns whether the counter is currently zero
	IsZero() (isZero bool)
	// Reset triggers the current channel and resets the WaitGroup
	Reset() (w2 *WaitGroupCh)
	// Wait()
	Waitable
	// “waitGroupCh_count…”
	fmt.Stringer
}

// Waiter is the interface implemented by
// wait-free-locked inspectable [WaitGroupCh]
type Waiter interface {
	// Add() Done() DoneBool()
	WaitedOn
	// Add() Ch() Count() Counts() IsZero() Reset() Wait()
	WaitingFor
}

// Doner implements awaitable exit for a goroutine
//   - implemented by:
//   - [parl.Go.Done]
//   - [parl.GoResult.Done]
type Doner interface {
	// Done indicates that this goroutine is exiting
	//	- err nil: successful exit
	//	- err non-nil: fatal error exit
	//	- —
	// 	- deferrable
	//   - Done makes a goroutine:
	//   - — awaitable and
	//   - — able to return error
	//   - — other needs of a goroutine is to initiate and detect cancel and
	//		submit non-fatal errors
	Done(errp *error)
}

// WaitLegacy is consumer methods compatible with [sync.WaitGroup]
type WaitLegacy interface {
	// Add()
	SyncAdd
	// Wait()
	Waitable
}

// DoneLegacy is the goroutine part of [sync.Waitgroup]
// and other implementations [WaitGroupCh]
//   - DoneLegacy is a many-to-many relation.
//   - DoneLegacy allows the callee to instatiate and invoke any number
//     of things that are awaitable by the caller.
//
// Usage:
//
//	… = NewSomething(&waitsForLots, &shutsDownLots)
//	go someThread(&waitsForLots, &shutsDownLots)
//	func someThread(DoneLegacy w, context.Context ctx) {
//	  defer w.Done()
//	  w.Add(2)
//	  go somethingElse()
type DoneLegacy interface {
	// Add()
	SyncAdd
	// Done()
	SyncDone
}

// Waitable is the invoker’s Wait part of [sync.WaitGroup.Wait] and
// compatible implementations: [WaitGroupCh.Wait].
//   - Waitable is a many-to-many relation
//   - Waitable allows a consumer creating goroutines to await exit of all threads
//
// Usage:
//
//	var waitsForLots parl.WaitGroupCh
//	var shutsDownLots parl.OnceChan
//	waitsForLots.Add(1)
//	go manyThreads(&waitsForLots, &shutsDownLots)
//	<-waitsForLots.Ch()
//	…
//	func manyThreads(Doneable w, context.Context ctx) {
//	  defer w.Done()
//	  w.Add(1)
//	  go otherThreads(w)
//	  <-ctx.Done()
//	  return
type Waitable interface {
	// Wait blocks until the WaitGroup counter is zero.
	Wait()
}

// SyncAdd is implemented by [sync.WaitGroup]
// and [WaitGroupCh]
type SyncAdd interface {
	// Add adds delta, which may be negative, to the WaitGroup counter.
	// If the counter becomes zero, all goroutines blocked waiting are released.
	// If the counter goes negative, Add panics.
	Add(delta int)
}

// SyncDone is implemengted by [sync.WaitGroup] and
// and [WaitGroupCh]
type SyncDone interface {
	// Done decrements the WaitGroup counter by one
	Done()
}
