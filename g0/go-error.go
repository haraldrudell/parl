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

type GoError struct {
	err        error
	t          time.Time
	errContext parl.GoErrorContext
	g0         parl.Go
}

func NewGoError(err error, errContext parl.GoErrorContext, g0 parl.Go) (goError parl.GoError) {
	return &GoError{
		err:        perrors.Stack(err),
		t:          time.Now(),
		errContext: errContext,
		g0:         g0,
	}
}

func (ge *GoError) Error() (message string) {
	if ge.err != nil {
		message = ge.err.Error()
	}
	return
}

func (ge *GoError) Time() (when time.Time) {
	return ge.t
}

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
		message = err.Error()
	} else {
		message = "OK"
	}
	return "error:\x27" + message + "\x27context:" + ge.errContext.String() + s
}
