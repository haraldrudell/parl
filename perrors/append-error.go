/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"github.com/haraldrudell/parl/internal/cyclebreaker2"
	"github.com/haraldrudell/parl/perrors/errorglue"
)

// AppendError associates an additional error with err.
//   - return value is nil if and only if both err and err2 are nil
//   - if either err or err2 is nil, return value is the non-nil argument
//   - if both err and err2 are non-nil, the return value is err with err2 associated
//   - associated error instances can be retrieved using:
//   - — perrors.AllErrors,
//   - — perros.ErrorList or
//   - — by rich error printing of perrors package: perrors.Long or
//   - — “%+v”
func AppendError(err error, err2 error) (e error) {
	if err2 == nil {
		e = err // err2 is nil, return is err, possibly nil
	} else if err == nil {
		e = err2 // err is nil, return is non-nil err2
	} else {
		e = errorglue.NewRelatedError(err, err2) // both non-nil
	}
	return
}

// ErrorList returns all error instances from a possible error chain.
// — If err is nil an empty slice is returned.
// — If err does not have associated errors, a slice of err, length 1, is returned.
// — otherwise, the first error of the returned slice is err followed by
//
//		other errors oldest first.
//	- Cyclic error values are dropped
func ErrorList(err error) (errs []error) {
	return errorglue.ErrorList(err)
}

// DeferredAppendError copies any error in errSource
// to errDest
//   - deferrable version of [AppendError]
//   - use case: single-threaded, deferred function
//     aggregating function-local error value to errp
func DeferredAppendError(errSource, errDest *error) {
	cyclebreaker2.NilPanic("errSource", errSource)
	cyclebreaker2.NilPanic("errDest", errDest)
	var err = *errSource
	if err == nil {
		return // no error
	}
	*errDest = AppendError(*errDest, err)
}
