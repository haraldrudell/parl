/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

// ErrorsWithStack gets all errors in the err error chain
// that has a stack trace.
// Oldest innermost stack trace is returned first.
// if not stack trace is present, the slice is empty
func ErrorsWithStack(err error) (errs []error) {
	for err != nil {
		if _, ok := err.(ErrorCallStacker); ok {
			errs = append([]error{err}, errs...)
		}
		err, _, _ = Unwrap(err)
	}
	return
}
