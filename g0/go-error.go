/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/errorglue"
	"github.com/haraldrudell/parl/perrors"
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
func (ge *GoError) Error() (message string) {
	if ge.err != nil {
		message = ge.err.Error()
	}
	return
}

// Time returns when the GoError was created
func (ge *GoError) Time() (when time.Time) {
	return ge.t
}

// Err returns the unbderlying error
func (ge *GoError) Err() (err error) {
	return ge.err
}

func (ge *GoError) IsThreadExit() (isThreadExit bool) {
	return ge.errContext == parl.GeExit ||
		ge.errContext == parl.GePreDoneExit
}

func (ge *GoError) IsFatal() (isThreadExit bool) {
	return (ge.errContext == parl.GeExit ||
		ge.errContext == parl.GePreDoneExit) &&
		ge.err != nil
}

func (ge *GoError) ErrContext() (errContext parl.GoErrorContext) {
	return ge.errContext
}

func (ge *GoError) Go() (g0 parl.Go) {
	return ge.g0
}

func (ge *GoError) String() (s string) {
	err := ge.err
	if stack := errorglue.GetInnerMostStack(err); len(stack) > 0 {
		s = "-at:" + stack[0].Short()
	}
	var message string
	if err != nil {
		message = perrors.Short(err)
	} else {
		message = "OK"
	}
	return "error:\x27" + message + "\x27context:" + ge.errContext.String() + s
}
