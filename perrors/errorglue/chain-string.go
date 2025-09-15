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

// ChainString() gets a string representation of a single error chain
//   - err: the error chain
//   - err nil: prints “OK”
//   - format: how error is printed
//   - — DefaultFormat: err.Error()
//   - — CodeLocation: one error: “message at runtime/panic.go:914”
//   - — ShortFormat: error and associated errors with code-location
//   - — ShortSuffix: no message “runtime/panic.go:914”
//   - s: printable string
func ChainString(err error, format CSFormat) (s string) {

	// no error case
	if err == nil {
		// no error return "OK"
		s = errorIsNilString
		return
	}

	switch format {
	case DefaultFormat:
		// like printf %v, printf %s and error.Error()
		//	- the first error in the chain has our error message
		s = err.Error()
		return
	case CodeLocation:
		// only one errror, with code location
		s = shortFormat(err)
		return
	case ShortFormat:
		// one-liner with code location and associated errors
		s = shortFormat(err)

		// add appended errors at the end 2[…]
		var list = ErrorList(err)
		if len(list) > 1 {
			// skip err itself
			list = list[1:]
			var sList = make([]string, len(list))
			for i, e := range list {
				sList[i] = shortFormat(e)
			}
			s += fmt.Sprintf(" %d[%s]", len(list), strings.Join(sList, ", "))
		}
		return
	case ShortSuffix:
		// for stackError, this provide the code-location without leading “ at ”
		s = codeLocation(err)
		return
	case LongFormat:
		// all errors with message and type
		//	- stack traces
		//	- related data
		//	- associated errors
	case LongSuffix:
		// first error of each error-chain in long format
		//	- an error chain is the initial error and any related errors
		//	- stack traces and data for all errors
	default:
		var stack = pruntime.NewStack(0)
		var packFuncS string
		if len(stack.Frames()) > 0 {
			packFuncS = stack.Frames()[0].Loc().PackFunc() + "\x20"
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

	// traverse all error instances
	//	- the initial error and any unique related error
	for i := 0; i < len(errorsToPrint); i++ {

		// every error is an error chain
		//	- traverse error chain
		var isFirstInChain = true
		for err = errorsToPrint[i]; err != nil; {

			// traverse the next node of the error chain
			var nextErr, joinedErrors, associatedError = Unwrap(err)

			// store associated errors
			if associatedError != nil {
				// add any new errors to errorsToPrint
				if !errorMap[associatedError] {
					errorMap[associatedError] = true
					errorsToPrint = append(errorsToPrint, associatedError)
				}
			}

			// store joinedErrors
			if len(joinedErrors) > 0 {
				errorsToPrint = append(errorsToPrint, joinedErrors...)
			}

			// ChainStringer errors produce their own representations
			var errorAsString string
			if richError, ok := err.(ChainStringer); ok {
				if isFirstInChain {
					errorAsString = richError.ChainString(LongFormat)
				} else {
					errorAsString = richError.ChainString(format)
				}
			} else {

				// regular errors
				//	- LongFormat prints all with type
				//	- LongSuffix only prints if first in chain
				if format == LongFormat || isFirstInChain {
					errorAsString = fmt.Sprintf("%s [%T]", err.Error(), err)
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
			err = nextErr
		}
	}

	return
}

const (
	// the string error value for error nil “OK”
	errorIsNilString = "OK"
)
