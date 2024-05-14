/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// Waitable is the invoker’s Wait part of sync.WaitGroup and
// other implementations.
// Waitable is a many-to-many relation.
// Waitable allows the caller to await exit and free of all invocations.
//
//	waitsForLots parl.WaitGroup
//	shutsDownLots parl.OnceChan
//	… = NewSomething(&waitsForLots, &shutsDownLots)
//	go someThread(&waitsForLots, &shutsDownLots)
//	func someThread(Doneable w, context.Context ctx) {
//	  defer w.Done()
//	  w.Add(2)
//	  go somethingElse()
type Waitable interface {
	Wait() // similar to sync.WaitGroup.Wait()
}

// SyncWait provides sync.WaitGroup.Wait()
type SyncWait interface {
	Wait()
}

// SyncWait provides sync.WaitGroup.Add()
type SyncAdd interface {
	Add(delta int)
}

// SyncDone provides sync.WaitGroup.Done()
type SyncDone interface {
	Done()
}

type WaitedOn interface {
	SyncAdd
	SyncDone
	DoneBool() (isExit bool)
	IsZero() (isZero bool)
}

type WaitingFor interface {
	SyncAdd
	IsZero() (isZero bool)
	Counters() (adds, dones int)
	SyncWait
	String() (s string)
}

// waiter allows to use any of observable parl.WaitGroup or parl.TraceGroup
type Waiter interface {
	WaitedOn
	WaitingFor
}

type ErrorManager interface {
	Ch() (ch <-chan GoError)
}

type ErrorCloser interface {
	InvokeIfError(addError func(err error))
	Close()
}

// Doneable is the callee part of sync.Waitgroup
// and other implementations
// Doneable is a many-to-many relation.
// Doneable allows the callee to instatiate and invoke any number
// of things that are awaitable by the caller.
//
//	… = NewSomething(&waitsForLots, &shutsDownLots)
//	go someThread(&waitsForLots, &shutsDownLots)
//	func someThread(Doneable w, context.Context ctx) {
//	  defer w.Done()
//	  w.Add(2)
//	  go somethingElse()
type Doneable interface {
	Add(delta int)
	Done()
}
