/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"errors"

	"github.com/haraldrudell/parl/perrors"
)

// RecoverInvocationPanic is intended to wrap callback invocations in the callee in order to
// recover from panics in the callback function.
// when an error occurs, perrors.AppendError appends the callback error to *errp.
// if fn is nil, a recovered panic results.
// if errp is nil, a panic is thrown, can be check with:
//
//	if errors.Is(err, parl.ErrNilValue) …
func RecoverInvocationPanic(fn func(), errp *error) {
	if errp == nil {
		panic(perrors.ErrorfPF("%w", ErrErrpNil))
	}
	defer Recover(Annotation(), errp, NoOnError)

	fn()
}

// ErrErrpNil indicates that a function with an error pointer argument
// received an errp nil value.
//
//	if errors.Is(err, parl.ErrNilValue) …
var ErrErrpNil = NilValueError(errors.New("errp cannot be nil"))

func RecoverInvocationPanicErr(fn func() (err error)) (isPanic bool, err error) {
	defer Recover(Annotation(), &err, NoOnError)

	isPanic = true
	err = fn()
	isPanic = false

	return
}

type TFunc[T any] func() (value T, err error)

type TResult[T any] struct {
	Value   T
	IsPanic bool
	Err     error
}

// RecoverInvocationPanicT invokes resolver, recover panics and populates v
func NewTResult[T any](tFunc TFunc[T]) (tResult *TResult[T]) {
	var t = TResult[T]{IsPanic: true}
	tResult = &t
	defer Recover(Annotation(), &t.Err, NoOnError)

	t.Value, t.Err = tFunc()
	t.IsPanic = false
	return
}
