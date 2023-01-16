/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import "errors"

// ErrorList returns all error instances from a possible error chain.
// — If err is nil an empty slice is returned.
// — If err does not have associated errors, a slice of err, length 1, is returned.
// — otherwise, the first error of the returned slice is err followed by
//
//		other errors oldest first.
//	- Cyclic error values are dropped
func ErrorList(err error) (errs []error) {
	if err == nil {
		return
	}
	err0 := err
	errMap := map[error]bool{err: true}
	for err != nil {
		if e, ok := err.(RelatedError); ok {
			if e2 := e.AssociatedError(); e2 != nil {
				if _, ok := errMap[e2]; !ok {
					errs = append([]error{e2}, errs...)
					errMap[e2] = true
				}
			}
		}
		err = errors.Unwrap(err)
	}
	return append([]error{err0}, errs...)
}
