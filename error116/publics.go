/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

import "github.com/haraldrudell/parl/errorglue"

// error116.Warning indicates that err is a problem of less severity than error.
// It is uesed for errors that are not to terminate the thread.
// A Warning can be detected using error116.IsWarning().
func Warning(err error) error {
	return Stack(errorglue.NewWarning(err))
}

// error116.AddKeyValue attaches a string value to err.
// The values can be trrioeved using error116.ErrorData().
// if key is non-empty valiue is returned in a map where last key wins.
// if key is empty, valuse is returned in s string slice.
// err can be nil.
func AddKeyValue(err error, key, value string) (e error) {
	return errorglue.NewErrorData(err, key, value)
}

// error116.AppendError associates an additional error with err.
// err and err2 can be nil.
// Associated error instances can be retrieved using error116.AllErrors, error116.ErrorList or
// by printing using rich error printing of the error116 package.
// TODO 220319 fill in printing
func AppendError(err error, err2 error) (e error) {
	if err2 == nil {
		return err // noop return
	}
	if err == nil {
		return err2 // single error return
	}
	return errorglue.NewRelatedError(err, err2)
}
