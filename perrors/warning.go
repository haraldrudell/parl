/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"errors"

	"github.com/haraldrudell/parl/perrors/errorglue"
)

// Warning indicates that err is a problem of less severity than error
//   - err: an error to be considered a warning
//   - err: nil: returns nil
//   - err2 is ensured to have a stack tgrace based on Warning caller
//   - Warning is detected using [IsWarning]
func Warning(err error) (err2 error) {
	if err == nil {
		return // err == nil → err2 == nil
	} else if !HasStack(err) {
		err = Stackn(err, skipOneStackFrame)
	}
	err2 = errorglue.NewWarning(err)
	return
}

// IsWarning determines if an error has been flagged as a warning
//   - isWarning true if err was wrapped by [Warning] function
func IsWarning(err error) (isWarning bool) {
	for ; err != nil; err = errors.Unwrap(err) {
		if _, isWarning = err.(*errorglue.WarningType); isWarning {
			return // is warning
		}
	}
	return // not a warning
}

const (
	// skip the current function in stack frame
	skipOneStackFrame = 1
)
