/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"errors"

	"github.com/haraldrudell/parl/perrors"
)

// ErrEndCallbacks indicates upon retun from a callback function that
// no more callbacks are desired. It does not indicate an error and is not returned
// as an error by any other function than the callback.
//
// callback invocations may be thread-safe, re-entrant and panic-handling but
// this deopends on the callback-invoking implementation used.
//
//	if errors.Is(err, parl.ErrEndCallbacks) { …
var ErrEndCallbacks = errors.New("end callbacks error")

// ErrErrpNil indicates that a function with an error pointer argument
// received an errp nil value.
//
//	if errors.Is(err, ErrErrpNil) …
var ErrErrpNil = errors.New("errp cannot be nil")

// RecoverInvocationPanic is intended to wrap callback invocations in the callee in order to
// recover from panics in the callback function.
// when an error occurs, perrors.AppendError appends the callback error to *errp.
// if fn is nil, a recovered panic results.
// if errp is nil, a panic is thrown, can be check with:
//
//	if errors.Is(err, ErrErrpNil) …
func RecoverInvocationPanic(fn func(), errp *error) {
	if errp == nil {
		panic(perrors.ErrorfPF("%w", ErrErrpNil))
	}
	defer Recover(Annotation(), errp, NoOnError)

	fn()
}
