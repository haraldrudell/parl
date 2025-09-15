/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import "github.com/haraldrudell/parl/perrors/errorglue"

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
