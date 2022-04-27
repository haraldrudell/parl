/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"fmt"
)

const (
	// DefaultFormat is similar to printf %v, printf %s and error.Error().
	// For an error with data, stack trace and associated errors,
	// DefaultFormat only prints the error message:
	//   error-message
	DefaultFormat CSFormat = iota + 1
	// ShortFormat has one-line location similar to printf %-v.
	// ShortFormat does not print stack traces, data and associated errors.
	// ShortFormat does print a one-liner of the error message and a brief code location:
	//   error-message at error116.(*csTypeName).FuncName-chainstring_test.go:26
	ShortFormat
	// LongFormat is similar to printf %+v.
	// ShortFormat does not print stack traces, data and associated errors.
	//   error-message
	//     github.com/haraldrudell/parl/error116.(*csTypeName).FuncName
	//       /opt/sw/privates/parl/error116/chainstring_test.go:26
	//     runtime.goexit
	//       /opt/homebrew/Cellar/go/1.17.8/libexec/src/runtime/asm_arm64.s:1133
	LongFormat
	// ShortSuffix one-line without message
	ShortSuffix
	// LongSuffix full stack trace without message
	LongSuffix
)

const csNilString = "OK"

// CSFormat describes string conversion of an error chain
type CSFormat byte

// ChainString() gets a string representation of a single error chain
// TODO 220319 finish comment
func ChainString(err error, format CSFormat) (s string) {
	if err == nil {
		return csNilString
	}
	if format == DefaultFormat {
		return err.Error() // the first error in the chain has our error message
	}
	if format == ShortFormat {
		// append the top frame of the oldest, innermost stack trace code location
		for _, e := range ErrorsWithStack(err) { // list with oldest first
			if e2, ok := e.(ChainStringer); ok {
				loc := e2.ChainString(ShortSuffix)
				if loc != "" {
					return err.Error() + loc
				}
			}
		}
		return err.Error() // no stack had a location available
	}

	// isIgnore: avoid cyclic traversal
	errorMap := map[error]bool{}
	isIgnore := func(err error) (ignore bool) {
		if errorMap[err] {
			return true
		}
		errorMap[err] = true
		return false
	}

	// errorsToPrint: list of discovered associated errors to print
	errorsToPrint := []error{err}
	addAnotherError := func(err error) {
		if isIgnore(err) {
			return
		}
		errorsToPrint = append(errorsToPrint, err)
	}

	// LongFormat
	// traverse error instances
	for i := 0; i < len(errorsToPrint); i++ {

		// traverse error chain
		isFirst := true
		for err = errorsToPrint[i]; err != nil; err = errors.Unwrap(err) {

			// look for associated errors
			if e2, ok := err.(RelatedError); ok {
				if e3 := e2.AssociatedError(); e3 != nil {
					addAnotherError(e3)
				}
			}

			var s2 string
			// ChainStringer errors produce their own representations
			if e2, ok := err.(ChainStringer); ok {
				if !isFirst && format == LongFormat {
					s2 = e2.ChainString(LongSuffix)
				} else {
					s2 = e2.ChainString(format)
				}
			} else {

				// for non-rich errors, only print the first one
				if isFirst {
					s2 = err.Error()
				}
			}

			if len(s2) > 0 {
				if len(s) > 0 {
					s += "\n" + s2
				} else {
					s = s2
				}
			}
			isFirst = false
		}
	}
	return
}

// PrintfFormat gets the ErrorFormat to use when executing
// the Printf value verb 'v'
func PrintfFormat(s fmt.State) CSFormat {
	if IsPlusFlag(s) {
		return LongFormat
	} else if IsMinusFlag(s) {
		return ShortFormat
	}
	return DefaultFormat
}
