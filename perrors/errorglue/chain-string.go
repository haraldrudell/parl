/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"fmt"
	"strings"

	"github.com/haraldrudell/parl/pruntime"
)

const (
	// string prepended to code location: “ at ”
	atStringChain = "\x20at\x20"
	// the string error value for error nil “OK”
	errorIsNilString = "OK"
)

// ChainString() gets a string representation of a single error chain
// TODO 220319 finish comment
func ChainString(err error, format CSFormat) (s string) {

	// no error case
	if err == nil {
		return errorIsNilString // no error return "OK"
	}

	switch format {
	case DefaultFormat: // like printf %v, printf %s and error.Error()
		s = err.Error() // the first error in the chain has our error message
		return
	case CodeLocation:
		s = shortFormat(err)
		return
	case ShortFormat: // one-liner with code location appended associated errors
		s = shortFormat(err)

		//add appended errors at the end 2[…]
		var list = ErrorList(err)
		if len(list) > 1 {
			list = list[1:] // skip err itself
			var sList = make([]string, len(list))
			for i, e := range list {
				sList[i] = shortFormat(e)
			}
			s += fmt.Sprintf(" %d[%s]", len(list), strings.Join(sList, ", "))
		}

		return
	case ShortSuffix:
		s = codeLocation(err)
		return
	case LongFormat, LongSuffix:
	default:
		var stack = pruntime.NewStackSlice(0)
		var packFuncS string
		if len(stack) > 0 {
			packFuncS = stack[0].PackFunc() + "\x20"
		}
		var e = fmt.Errorf("%sbad format: %s", packFuncS, format)
		panic(NewErrorStack(e, stack))
	}
	// LongFormat, LongSuffix: recursive printing and associated errors

	// errorMap is a map of errors already printed
	//	- it is used to avoid cyclic printing
	var errorMap = map[error]bool{}
	// errorsToPrint: list of discovered associated errors to print
	var errorsToPrint = []error{err}

	// traverse error instances
	for i := 0; i < len(errorsToPrint); i++ {

		// traverse error chain
		var isFirstInChain = true
		for err = errorsToPrint[i]; err != nil; err = errors.Unwrap(err) {

			// look for associated errors
			if e2, ok := err.(RelatedError); ok {
				if e3 := e2.AssociatedError(); e3 != nil {
					// add any new errors to errorsToPrint
					if !errorMap[e3] {
						errorMap[e3] = true
						errorsToPrint = append(errorsToPrint, e3)
					}
				}
			}

			// ChainStringer errors produce their own representations
			var errorAsString string
			if richError, ok := err.(ChainStringer); ok {
				if !isFirstInChain && format == LongFormat {
					errorAsString = richError.ChainString(LongSuffix)
				} else {
					errorAsString = richError.ChainString(format)
				}
			} else {

				// for non-rich errors, only print the first one
				if isFirstInChain {
					errorAsString = err.Error()
				}
			}

			if len(errorAsString) > 0 {
				if len(s) > 0 {
					s += "\n" + errorAsString
				} else {
					s = errorAsString
				}
			}
			isFirstInChain = false
		}
	}

	return
}

// shortFormat: “message at runtime/panic.go:914”
//   - if err or its error-chain does not have location: “message” like [error.Error]
//   - if err or its error-chain has panic, location is the code line
//     that caused the first panic
//   - if err or its error-chain has location but no panic,
//     location is where the oldest error with stack was created
//   - err is non-nil
func shortFormat(err error) (s string) {

	// append the top frame of the oldest, innermost stack trace code location
	s = codeLocation(err)
	if s != "" {
		s = err.Error() + atStringChain + s
	} else {
		s = err.Error()
	}

	return
}

// codeLocation: “runtime/panic.go:914”
//   - if err or its error-chain does not have location: empty string
//   - if err or its error-chain has panic, location is the code line
//     that caused the first panic
//   - if err or its error-chain has location but no panic,
//     location is where the oldest error with stack was created
//   - err is non-nil, no “ at ” prefix
func codeLocation(err error) (message string) {

	// err or err’s error-chain may contain stacks
	//	- any of the stacks may contain a panic
	//	- an error with stack is able to locate any panic it or its chain has
	//	- therefore scan for any error with stack and ask the first one for location
	for e := err; e != nil; e = errors.Unwrap(e) {
		if _, ok := e.(ErrorCallStacker); !ok {
			continue // e does not have stack
		}
		var _ = (&errorStack{}).ChainString
		message = e.(ChainStringer).ChainString(ShortSuffix)
		return // found location return
	}

	return // no location return
}

// PrintfFormat gets the ErrorFormat to use when executing
// the Printf value verb 'v'
//   - %+v: [LongFormat]
//   - %-v: [ShortFormat]
//   - %v: [DefaultFormat]
func PrintfFormat(s fmt.State) CSFormat {
	if IsPlusFlag(s) {
		return LongFormat
	} else if IsMinusFlag(s) {
		return ShortFormat
	}
	return DefaultFormat
}
