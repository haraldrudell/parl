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
type Future[T any] struct {
	await Awaitable // one-to-many wait mechanic based on channel
	// result holds any successful result of calculation
	//	- result value is the time calculating began
	//	- result holds zero-value if the calculation failed
	//	- result is updated by the winner thread prior to lock.Unlock
	result atomic.Pointer[TResult[T]]
}

// NewFuture returns an awaitable calculation
func NewFuture[T any]() (calculation *Future[T]) { return &Future[T]{await: *NewAwaitable()} }

// IsCompleted returns whether the calculation is complete. Thread-safe
func (f *Future[T]) IsCompleted() (isCompleted bool) { return f.await.IsClosed() }

// Ch returns an awaitable channel
func (f *Future[T]) Ch() (ch AwaitableCh) { return f.await.Ch() }

// Result retrieves the calculation’s result. May block. Thread-safe
func (f *Future[T]) Result() (result T, isValid bool) {
	<-f.await.Ch()
	if rp := f.result.Load(); rp != nil {
		result = rp.Value
		isValid = rp.Err == nil
	}
	return
}

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
