/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import "errors"

// Error0 returns the last error in err’s error chain
// or nil if err is nil
func Error0(err error) (e error) {
	for ; err != nil; err = errors.Unwrap(err) {
		e = err
	}
	return
}
