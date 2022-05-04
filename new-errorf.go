/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/perrors"

// parl.Errorf offers stack traces and other rich error featues of
// packages perrors and errorglue
func Errorf(format string, a ...interface{}) (err error) {
	return perrors.Errorf(format, a...)
}

// parl.New offers stack traces and other rich error featues of
// packages perrors and errorglue
func New(s string) error {
	return perrors.New(s)
}
