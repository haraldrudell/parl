/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"

	"github.com/haraldrudell/parl/perrors"
)

const (
	// counts the stack frame of [parl.ensureError]
	parlEnsureErrorFrames = 1
	// counts the stack frame of [parl.EnsureError]
	parlEnsureErrorFrames0 = 1
)

// ensureError interprets a panic values as an error
//   - returned value is either nil or an error value with stack trace
//   - the error is ensured to have stack trace
func EnsureError(panicValue any) (err error) {
	return ensureError(panicValue, parlEnsureErrorFrames0)
}

// ensureError interprets a panic values as an error
//   - returned value is either nil or an error value with stack trace
//   - frames is used to select the stack frame from where the stack trace begins
//   - frames 0 is he caller of ensure Error
func ensureError(panicValue any, frames int) (err error) {

	// no panic is no-op
	if panicValue == nil {
		return // no panic return
	}

	// ensure value to be error
	var ok bool
	if err, ok = panicValue.(error); !ok {
		err = fmt.Errorf("non-error value: %T %+[1]v", panicValue)
	}

	// ensure stack trace
	if !perrors.HasStack(err) {
		if frames < 0 {
			frames = 0
		}
		err = perrors.Stackn(err, frames+parlEnsureErrorFrames)
	}

	return
}
