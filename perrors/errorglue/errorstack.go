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

func (errorStackValue *errorStack) ChainString(format CSFormat) (s string) {
	// format can be DefaultFormat ShortFormat LongFormat ShortSuffix LongSuffix
	if errorStackValue == nil {
		return // nil error return: no location

		// error message into s
	} else if format != ShortSuffix {
		s = errorStackValue.Error() // all formats less ShortSuffic feature the error message
		if format == DefaultFormat {
			return // DefaultFormat return: error message, no location
		}
	}

	// LongFormat or LongSuffix: full stack trace
	if format == LongSuffix || format == LongFormat {
		return s + errorStackValue.s.String() // long: error message followed by the entire stack trace
	}

	// it is ShortFormat or ShortSuffix: find the code location where the error occured
	// code location
	var shortCodeLocationString string
	// special case for panics: find the true panic location
	if isPanic, _, panicIndex := Indices(errorStackValue.s); isPanic {
		// panic-causing location
		shortCodeLocationString = errorStackValue.s[panicIndex].Short()
	} else {
		// regular code location
		shortCodeLocationString = errorStackValue.s.Short()
	}

	// ShortFormat: error message and location with " at " in between
	// ShortSuffix: no leading " at "
	// some stack traces have " at ", others do not
	if format == ShortFormat {
		// ensure stack frame begins with " at "
		if !strings.HasPrefix(shortCodeLocationString, errorStackAtString) {
			shortCodeLocationString = errorStackAtString + shortCodeLocationString
		}
		return s + shortCodeLocationString
	}

	// ShortSuffix: ensure no " at "
	return strings.TrimPrefix(shortCodeLocationString, errorStackAtString)
}
