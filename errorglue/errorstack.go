/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"fmt"

	"github.com/haraldrudell/parl/pruntime"
)

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

func (e *errorStack) ChainString(format CSFormat) (s string) {
	if e == nil {
		return
	} else if format != ShortSuffix {
		s = e.Error()
		if format == DefaultFormat {
			return
		}
	}
	if format == ShortSuffix || format == ShortFormat {
		return s + e.s.Short()
	}

	// LongFormat or LongSuffix
	return s + e.s.String()
}
