/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"errors"

	"github.com/haraldrudell/parl/perrors/errorglue"
)

// HasStack detects if the error chain already contains a stack trace
//   - hasStack: true if err is non-nil and contains a stack trace
func HasStack(err error) (hasStack bool) {
	if err == nil {
		return
	}
	var e errorglue.ErrorCallStacker
	// if an error of type ErrorCallStacker is found in err’s error chain,
	// hasStack is true
	hasStack = errors.As(err, &e)
	return
}
