/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"fmt"
	"strings"

	"github.com/haraldrudell/parl/pruntime"
)

// errorStack provides a stack trace to an error chain.
//   - errorStack has publics Error() Unwrap() Format() ChainString() StackTrace()
type errorStack struct {
	// Format() Unwrap() Error()
	RichError
	s pruntime.Stack
}

// errorStack implements [error]
var _ error = &errorStack{}

// errorStack implements [ErrorCallStacker] can provide a stack trace [ErrorCallStacker.StackTrace]
var _ ErrorCallStacker = &errorStack{}

// errorStack implements [ChainStringer]: [ChainStringer.ChainString]
var _ ChainStringer = &errorStack{}

// errorStack implements [fmt.Formatter]: has features for [fmt.Printf]: [fmt.Formatter.Format]
var _ fmt.Formatter = &errorStack{}

// errorStack implements [Wrapper]: has an error chain [Unwrapper.Unwrap]
var _ Unwrapper = &errorStack{}

// NewErrorStack attaches a stack to err
func NewErrorStack(err error, st pruntime.Stack) (e2 error) {
	return &errorStack{*newRichError(err), st}
}

// StackTrace returns a clone of e’s stack trace
func (e *errorStack) StackTrace() (st pruntime.Stack) {
	if e == nil {
		return
	}
	return e.s
}

// ChainString is invoked by [ChainStringer] to get a specific format
//   - nil error or unknown format returns empty string “”
//   - DefaultFormat: “message”
//   - ShortFormat: “message at runtime/panic.go:914”
//   - LongFormat: “message \n runtime.gopanic \n runtime/panic.go:914”
//   - ShortSuffix: “runtime/panic.go:914”
//   - LongSuffix: “message \n runtime.gopanic \n runtime/panic.go:914”
func (errorStackValue *errorStack) ChainString(format CSFormat) (s string) {
	if errorStackValue == nil {
		return // nil error return: no location
	}
	switch format {
	case DefaultFormat:
		s = errorStackValue.Error()
		return
	case ShortFormat:
		_, s = errorStackValue.shortCodeLocationString()
		// ensure stack frame begins with " at "
		if !strings.HasPrefix(s, errorStackAtString) {
			s = errorStackAtString + s
		}
		s = errorStackValue.Error() + s
		return
	case ShortSuffix:
		_, s = errorStackValue.shortCodeLocationString()
		// ensure no " at "
		s = strings.TrimPrefix(s, errorStackAtString)
		return
	case LongFormat: // full stack trace
		// add location if trace contains panic
		if isPanic, loc := errorStackValue.shortCodeLocationString(); isPanic {
			s = loc
			if !strings.HasPrefix(s, errorStackAtString) {
				s = errorStackAtString + s
			}
		}
		s = fmt.Sprintf("%s [%T]%s\n%s",
			errorStackValue.Error(), errorStackValue, // “error-message [errors.Type]”
			s,                          // “ at runtime.gopanic:17”
			errorStackValue.s.String(), // multiple-line stack-trace
		)
		return
	case LongSuffix:
		s = errorStackValue.s.String()
		return
	default:
		return // unknown format
	}
}

// find the code location where the error occured
//   - “mains.(*Executable).AddErr-executable.go:25”
func (errorStackValue *errorStack) shortCodeLocationString() (isPanic bool, shortCodeLocationString string) {

	// check for a panic which will be the code location
	//	- instead of returning random runtime locations,
	//		this returns the code line where the panic occurred
	//	- this may be a stack from subordinate error in the error chain
	var frame pruntime.Frame
	if isPanic0, stack, _, panicIndex, _, _ := FirstPanicStack(errorStackValue); isPanic0 {
		isPanic = true
		var frames = stack.Frames()
		// code line that caused panic
		frame = frames[panicIndex]
	} else {
		// regular code location
		frame = errorStackValue.s.Frames()[0]
	}
	shortCodeLocationString = frame.Loc().Short()

	return
}

const (
	errorStackAtString = "\x20at\x20"
)
