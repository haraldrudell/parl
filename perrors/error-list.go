/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import "github.com/haraldrudell/parl/perrors/errorglue"

// ErrorList returns all error-chains contained in err
//   - err: the error to traverse. If err is nil an empty slice is returned.
//     Otherwise, err is the first error of the returned slice followed by
//     other error-chains oldest first.
//   - — If err does not have associated errors, a slice of error length 1 is returned
//   - —
//   - error chains are created by [errors.Join] or [perrors.AppendError]
//   - like errors.Is but supports [perrors.AppendError]
//   - Cyclic error values are dropped
//
// Usage:
//
//	for _, anError := range perrors.ErrorList(err) {
//	  if errors.Is(anError, context.Canceled) {
func ErrorList(err error) (errs []error) { return errorglue.ErrorList(err) }
