/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"github.com/haraldrudell/parl/perrors"
)

// FutureCore implements a future with value and error obtain in the Wait method. Thread-safe
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
type FutureCore[T any] struct {
	valueWait   WaitGroup
	futureValue FutureValue[T]
}

// FutureValue is a container for the value of the future.
type FutureValue[T any] struct {
	value   T     // happens-before via wg
	isPanic bool  // wether a panic occurred in fn
	err     error // happens-before via wg
}

type Resolver[T any] func() (value T, err error)

// NewFutureCore executes fn and returns a future for its result.
// Consider using WInOrWaiter mechanic instead.
// Usage:
//
//	f := NewFutureCore(computeTfunc)
//	value, isPanic, err := f.Wait()
func NewFutureCore[T any](resolver Resolver[T]) (futureCore *FutureCore[T]) {
	f := FutureCore[T]{}
	InitFutureCore(&f, resolver)
	return &f
}

// InitFutureCore initializes a future and executes fn
func InitFutureCore[T any](f *FutureCore[T],
	resolver Resolver[T],
) {
	if f == nil {
		panic(perrors.NewPF("f cannot be nil"))
	}
	if resolver == nil {
		panic(perrors.NewPF("resolver cannot be nil"))
	}
	f.valueWait.Add(1)
	go f.invoke(resolver)
}

// IsDone returns whether the future result is present
func (f *FutureCore[T]) IsDone() (isDone bool) {
	return f.valueWait.IsZero()
}

// Wait block until the result is available and returns it
func (f *FutureCore[T]) Wait() (value T, isPanic bool, err error) {
	f.valueWait.Wait()
	value = f.futureValue.value
	isPanic = f.futureValue.isPanic
	err = f.futureValue.err
	return
}

func (f *FutureCore[T]) invoke(fn func() (value T, err error)) {
	defer f.valueWait.Done()
	isPanic := true
	defer func() {
		f.futureValue.isPanic = isPanic
	}()
	defer Recover(Annotation(), &f.futureValue.err, NoOnError)

	f.futureValue.value, f.futureValue.err = fn()
	isPanic = false
}
