/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"errors"

	"github.com/haraldrudell/parl/perrors/errorglue"
	"github.com/haraldrudell/parl/pruntime"
)

// Stack ensures the err has a stack trace
// associated.
//   - err nil returns nil
func Stack(err error) (err2 error) {
	if HasStack(err) {
		err2 = err
		return
	}
	err2 = Stackn(err, e116StackFrames)
	return
}

// Stackn always attaches a new stack trace to non-nil err
//   - framesToSkip: 0 is caller, larger skips stack frames
func Stackn(err error, framesToSkip int) (err2 error) {
	if err == nil {
		return
	} else if framesToSkip < 0 {
		framesToSkip = 0
	}
	err2 = errorglue.NewErrorStack(
		err,
		pruntime.NewStack(e116StacknFramesToSkip+framesToSkip),
	)
	return
}

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
