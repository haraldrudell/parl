/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

type Future[T any] struct {
	threadWait WaitGroup
	futureCore FutureCore[T]
}

// NewFuture computes a future value in a separate goroutine.
// fn must be thread-safe.
// an fn panic is recovered.
// Result is either via Wait() or received from a channel from Ch().
func NewFuture[T any](fn func() (value T, err error)) (future *Future[T]) {
	f := Future[T]{}
	InitFutureCore(fn, &f.futureCore)
	return &f
}

func (f *Future[T]) Ch() (ch <-chan FutureValue[T]) {
	ch0 := make(chan FutureValue[T])
	ch = ch0
	f.threadWait.Add(1)
	go f.chThread(ch0)
	return
}

func (f *Future[T]) chThread(ch chan<- FutureValue[T]) {
	defer f.threadWait.Done()
	defer Recover(Annotation(), nil, Infallible)

	value, err := f.futureCore.Wait()
	ch <- FutureValue[T]{value: value, err: err}
}

func (f *Future[T]) IsDone() (isDone bool) {
	return f.futureCore.IsDone() && f.threadWait.IsZero()
}

func (f *Future[T]) Wait() (value T, err error) {
	value, err = f.futureCore.Wait()
	f.threadWait.Wait()
	return
}
