/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
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
