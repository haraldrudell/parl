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
//   - Future allows a thread to await a value calculated in parallel
//     by other threads
//   - unlike for a promise, consumer manages any thread,
//     therefore producing debuggable code and meaningful stack traces
//   - a promise launches the thread why there is no trace of what code
//     created the promise or why
type Future[T any] struct {
	// one-to-many wait mechanic based on channel
	await Awaitable
	// calculation outcome
	result atomic.Pointer[TResult[T]]
}

// NewFuture returns an awaitable calculation
//   - has an Awaitable and a thread-safe TResult container
//
// Usage:
//
//	 var calculation = NewFuture[someType]()
//	 go calculateThread(calculation)
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
func NewFuture[T any]() (calculation *Future[T]) { return &Future[T]{await: *NewAwaitable()} }

// IsCompleted returns whether the calculation is complete. Thread-safe
func (f *Future[T]) IsCompleted() (isCompleted bool) { return f.await.IsClosed() }

// Ch returns an awaitable channel. Thread-safe
func (f *Future[T]) Ch() (ch AwaitableCh) { return f.await.Ch() }

// Result retrieves the calculation’s result
//   - May block. Thread-safe
func (f *Future[T]) Result() (result T, hasValue bool) {

	// blocks here
	<-f.await.Ch()

	if rp := f.result.Load(); rp != nil {
		result = rp.Value
		hasValue = rp.Err == nil
	}

	return
}

// TResult returns a pointer to the future’s result
//   - nil if future has not resolved
//   - thread-safe
func (f *Future[T]) TResult() (tResult *TResult[T]) { return f.result.Load() }

// End writes the result of the calculation, deferrable
//   - value is considered valid if errp is nil or *errp is nil
//   - End can make a goroutine channel-awaitable
//   - End can only be invoked once or panic
//   - any argument may be nil
//   - thread-safe
func (f *Future[T]) End(value *T, isPanic *bool, errp *error) {

	// create result to swap-in for atomic
	var result = NewTResult3(value, isPanic, errp)

	// check for multiple invocations
	if !f.result.CompareAndSwap(nil, result) {
		panic(perrors.NewPF("End invoked multiple times"))
	}

	// trigger awaitable
	f.await.Close()
}
