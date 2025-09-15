/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

// ErrorChainSlice returns a slice of errors from following
// the main error chain
//   - err: an error to traverse
//   - errs: all errors in the main error chain beginning with err itself
//   - — nil if err was nil
//   - — otherwise of length 1 or more
func ErrorChainSlice(err error) (errs []error) {
	for err != nil {
		errs = append(errs, err)
		err, _, _ = Unwrap(err)
	}
	return
}
