/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"github.com/haraldrudell/parl/perrors"
)

// Promise is a future computed by a separate thread in parallel with the consumer. Thread-Safe
//   - any number of threads can use IsDone and Wait methods
//   - the function fn calculating the result must be thread-safe
//   - — fn execution is in a new thread
//   - Wait waits for fn to complete and returns its value and error
//   - IsDone checks whether the value is present
//   - FutureValue is in a separate type so that it can be sent on a channel
//
// Note: because the future value is computed in a new thread, tracing the sequence of events
// may prove difficult would the future cause partial deadlock by the resolver being blocked.
// The difficulty is that there is no continuous stack trace or thread ID showing what initiated the future.
// Consider using WinOrWaiter mechanic.
type Promise[T any] struct {
	// must be pointer or memory leak will result
	*Future[TResult[T]]
}

// NewPromise executes resolver in a new thread and returns a future for its result.
//   - Consider using WinOrWaiter mechanic
//
// Usage:
//
//	f := NewPromise(computeTfunc)
//	value, isPanic, err := f.Wait()
func NewPromise[T any](resolver TFunc[T], g0 Go) (promise *Promise[T]) {
	if resolver == nil {
		panic(perrors.NewPF("resolver cannot be nil"))
	}

	// launch computational thread
	future := *NewFuture[TResult[T]]()
	go promiseThread(resolver, &future, g0)

	return &Promise[T]{Future: &future}
}

// promiseThread is a thread that invokes resolver and future.End
func promiseThread[T any](resolver TFunc[T], future *Future[TResult[T]], g0 Go) {
	var err error
	defer g0.Done(&err)
	defer PanicToErr(&err)

	var promiseValue TResult[T]
	defer future.End(&promiseValue, &promiseValue.Err)

	promiseValue = *NewTResult(resolver)
}
