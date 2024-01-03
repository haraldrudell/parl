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

const errorStackAtString = "\x20at\x20"

// errorStack provides a stack trace to an error chain.
// errorStack has publics Error() Unwrap() Format() ChainString() StackTrace()
type errorStack struct {
	RichError
	s pruntime.StackSlice // slice
}

var _ error = &errorStack{}            // errorStack behaves like an error
var _ ErrorCallStacker = &errorStack{} // errorStack can provide a stack trace
var _ ChainStringer = &errorStack{}    // errorStack can be used by ChainString
var _ fmt.Formatter = &errorStack{}    // errorStack has features for fmt.Printf
var _ Wrapper = &errorStack{}          // errorStack has an error chain

func NewErrorStack(err error, st pruntime.StackSlice) (e2 error) {
	return &errorStack{*newRichError(err), st}
}

func (e *errorStack) StackTrace() (st pruntime.StackSlice) {
	if e == nil {
		return
	}
	return e.s.Clone()
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
		s = errorStackValue.Error() + s + errorStackValue.s.String()
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
	//		this returns the code line wherte the panic occurred
	//	- this may be a stack from subordinate error in the error chain
	if isPanic0, stack, _, panicIndex, _, _ := FirstPanicStack(errorStackValue); isPanic0 {
		isPanic = true
		// code line that caused panic
		shortCodeLocationString = stack[panicIndex].Short()
	} else {
		// regular code location
		shortCodeLocationString = errorStackValue.s.Short()
	}

	return
}
