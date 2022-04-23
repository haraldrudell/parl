/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "context"

// Waitable is the invoker’s Wait part of sync.WaitGroup and
// other implementations.
// Waitable is a many-to-many relation.
// Waitable allows the caller to await exit and free of all invocations.
//  waitsForLots parl.WaitGroup
//  shutsDownLots parl.OnceChan
//  … = NewSomething(&waitsForLots, &shutsDownLots)
//  go someThread(&waitsForLots, &shutsDownLots)
//  func someThread(Doneable w, context.Context ctx) {
//    defer w.Done()
//    w.Add(2)
//    go somethingElse()
type Waitable interface {
	Wait()
}

// Doneable is the callee part of sync.Waitgroup
// and other implementations
// Doneable is a many-to-many relation.
// Doneable allows the callee to instatiate and invoke any number
// of things that are awaitable by the caller.
//  … = NewSomething(&waitsForLots, &shutsDownLots)
//  go someThread(&waitsForLots, &shutsDownLots)
//  func someThread(Doneable w, context.Context ctx) {
//    defer w.Done()
//    w.Add(2)
//    go somethingElse()
type Doneable interface {
	Add(delta int)
	Done()
}

// context.WithCancel provides a thread-safe race-free idempotent
// many-to-many shutdown relation.
type CancelContext interface {
	context.Context
	Cancel()
}
