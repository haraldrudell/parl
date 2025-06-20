/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
)

// Future is a container for an awaitable calculation result
//   - used to complete a calculation in a separate thread.
//     Future is a calculation that is internally awaitable.
//     Usable as a task for worker threads.
//     Alternative is to pass values in channels
//   - [Future.Result] blocking wait for result
//   - [Future.Ch] awaitable channel
//   - [Future.IsCompleted] indicates if End was invoked
//   - [Future.End] calculation thread provides result
//   - Initialization-free low-allocation thread-safe
//   - contains non-pointer atomics: cannot be copied
//   - —
//   - has an Awaitable and a thread-safe TResult container
//   - performance: Future has internal T storage to avoid allocation.
//     A channel may be allocated by Awaitable.
//     If T contains non-pointer lock or atomic, T should be pointer.
//     Always allocated on stack
//   - Future allows a thread to await a value calculated in parallel
//     by other threads
//   - unlike for a promise, consumer manages any thread,
//     therefore producing debuggable code and meaningful stack traces
//   - a promise launches the thread why there is no trace of what code
//     created the promise or why
//
// Usage:
//
//	 var calculation Future[someType]
//	 go calculateThread(&calculation)
//	 …
//	 var result, isValid = calculation.Result()
//
//	func calculateThread(future *Future[someType]) {
//	  var err error
//	  var isPanic bool
//	  var value someType
//	  defer calculation.End(&value, &isPanic, &err)
//	  defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err, &isPanic)
//
//	   value = …
type Future[T any] struct {
	// one-to-many wait mechanic based on channel
	await Awaitable
	// calculation outcome storage
	//	- fields: Value IsPanic Err
	result TResult[T]
	// isEnd ensures single End invocation
	isEnd atomic.Uint32
}

// IsCompleted returns whether the calculation is complete. Thread-safe
func (f *Future[T]) IsCompleted() (isCompleted bool) { return f.await.IsClosed() }

// Ch returns an awaitable channel. Thread-safe
func (f *Future[T]) Ch() (ch AwaitableCh) { return f.await.Ch() }

// Result retrieves the calculation’s result
//   - May block until [Future.End] invocation
//   - resultp: valid if hasValue true.
//     Allocation is part of Future struct
//   - hasValue true: success, resultp available
//   - hasValue false: failure: [Future.TResult] returns error
//   - Thread-safe idempotent
func (f *Future[T]) Result() (resultp *T, hasValue bool) {

	// blocks here
	<-f.await.Ch()

	resultp = &f.result.Value
	hasValue = f.result.Err == nil

	return
}

// TResult returns a pointer to the future’s result
//   - nil if future has not resolved
//   - thread-safe idempotent
func (f *Future[T]) TResult() (tResult *TResult[T]) {
	if f.await.IsClosed() {
		tResult = &f.result
	}
	return
}

// End writes the result of the calculation, deferrable
//   - value: valid if errp is nil or *errp is nil. May be nil
//   - isPanic: indicates that the error was caused by panic
//     Valid if *errp is non-nil. May be [NoIsPanic] or nil
//   - errp: points to possible error. May be [NoErrp] or nil
//   - —
//   - End can make a goroutine channel-awaitable
//   - End can only be invoked once or panic
//   - thread-safe deferrable invoke-once
func (f *Future[T]) End(value *T, isPanic *bool, errp *error) {

	// gate for only one invocation 5.6 ns
	if f.isEnd.Load() != 0 || f.isEnd.Swap(1) != 0 {
		panic(perrors.NewPF("End invoked multiple times"))
	}

	NewTResult3(value, isPanic, errp, &f.result)

	// trigger awaitable
	f.await.Close()
}
