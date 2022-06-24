/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"github.com/haraldrudell/parl/perrors"
)

type FutureCore[T any] struct {
	valueWait   WaitGroup
	futureValue FutureValue[T]
}

type FutureValue[T any] struct {
	value T     // happens-before via wg
	err   error // happens-before via wg
}

func NewFutureCore[T any](fn func() (value T, err error)) (futureCore *FutureCore[T]) {
	f := FutureCore[T]{}
	InitFutureCore(fn, &f)
	return &f
}

func InitFutureCore[T any](fn func() (value T, err error), f *FutureCore[T]) (futureCore *FutureCore[T]) {
	if fn == nil {
		panic(perrors.NewPF("fn nil"))
	}
	f.valueWait.Add(1)
	go f.invoke(fn)
	return f
}

func (f *FutureCore[T]) IsDone() (isDone bool) {
	return f.valueWait.IsZero()
}

func (f *FutureCore[T]) Wait() (value T, err error) {
	f.valueWait.Wait()
	value = f.futureValue.value
	err = f.futureValue.err
	return
}

func (f *FutureCore[T]) invoke(fn func() (value T, err error)) {
	defer f.valueWait.Done()
	defer Recover(Annotation(), &f.futureValue.err, NoOnError)

	f.futureValue.value, f.futureValue.err = fn()
}
