/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/perrors/errorglue"
)

// GoError is a wrapper around an error associating it with a Go goroutine
// and the situation in which this error occurred
type GoError struct {
	err        error // err is the underlying unadulteraded error. It is nil for non-fatal Go exits
	t          time.Time
	errContext parl.GoErrorContext // errContext describes in what situation the error occured
	g0         parl.Go             // all errors are associated with a Go.
}

// NewGoError creates a GoError based on an error
func NewGoError(err error, errContext parl.GoErrorContext, g0 parl.Go) (goError parl.GoError) {
	return &GoError{
		err:        perrors.Stack(err),
		t:          time.Now(),
		errContext: errContext,
		g0:         g0,
	}
}

// Error returns a human-readable error message making GoError implement error
//   - for nil errors, empty string is returned
func (e *GoError) Error() (message string) {
	if e.err != nil {
		message = e.err.Error()
	}
	return
}

// Time returns when the GoError was created
func (e *GoError) Time() (when time.Time) {
	return e.t
}

// Err returns the underlying error
func (e *GoError) Err() (err error) {
	return e.err
}

// ErrString returns string representation of error
//   - if no error “OK”
//   - if not debug or panic, short error with location
//   - otherwise error with stack trace
func (e *GoError) ErrString() (errString string) {
	var isLong = parl.IsThisDebug()
	if !isLong {
		isLong, _, _, _ = perrors.IsPanic(e)
	}
	if isLong {
		errString = perrors.Long(e)
	} else {
		errString = perrors.Short(e)
	}
	return
}

func (e *GoError) IsThreadExit() (isThreadExit bool) {
	return e.errContext == parl.GeExit ||
		e.errContext == parl.GePreDoneExit
}

func (e *GoError) IsFatal() (isThreadExit bool) {
	return (e.errContext == parl.GeExit ||
		e.errContext == parl.GePreDoneExit) &&
		e.err != nil
}

func (e *GoError) ErrContext() (errContext parl.GoErrorContext) {
	return e.errContext
}

func (e *GoError) Go() (g0 parl.Go) {
	return e.g0
}

func (e *GoError) String() (s string) {
	var err = e.err
	var stack = errorglue.GetInnermostStack(err)
	if stack != nil {
		s = "-at:" + stack.Frames()[0].String()
	}
	var message string
	if err != nil {
		message = perrors.Short(err)
	} else {
		message = "OK"
	}
	return "error:\x27" + message + "\x27context:" + e.errContext.String() + s
}
