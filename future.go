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
//   - unlike for a promise, consumer manages any threads,
//     therefore debuggable and meaningful stack traces
type Future[T any] struct {
	await  Awaitable                  // one-to-many wait mechanic based on channel
	result atomic.Pointer[TResult[T]] // calculation outcome
}

// NewFuture returns an awaitable calculation
//   - has an Awaitable and a thread-safe TResult container
func NewFuture[T any]() (calculation *Future[T]) { return &Future[T]{await: *NewAwaitable()} }

// IsCompleted returns whether the calculation is complete. Thread-safe
func (f *Future[T]) IsCompleted() (isCompleted bool) { return f.await.IsClosed() }

// Ch returns an awaitable channel
func (f *Future[T]) Ch() (ch AwaitableCh) { return f.await.Ch() }

// Result retrieves the calculation’s result
//   - May block. Thread-safe
func (f *Future[T]) Result() (result T, isValid bool) {
	<-f.await.Ch()
	if rp := f.result.Load(); rp != nil {
		result = rp.Value
		isValid = rp.Err == nil
	}
	return
}

// TResult returns a pointer to the future’s result
//   - nil if future has not resolved
func (f *Future[T]) TResult() (tResult *TResult[T]) { return f.result.Load() }

// End writes the result of the calculation, deferrable
//   - value is considered valid if errp is nil or *errp is nil
//   - End can only be invoked once
//   - value isPanic errp can be nil
func (f *Future[T]) End(value *T, isPanic *bool, errp *error) {
	var result = NewTResult3(value, isPanic, errp)
	if !f.result.CompareAndSwap(nil, result) {
		panic(perrors.NewPF("End invoked multiple times"))
	}
	f.await.Close()
}
