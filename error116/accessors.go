/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

import (
	"errors"

	"github.com/haraldrudell/parl/errorglue"
)

// error116.ErrorData get possible string values associated with an error chain.
// list is a list of string values that were stored with an empty key, oldest first.
// keyValues are string values associated with a key string, newest key wins.
// err can be nil
func ErrorData(err error) (list []string, keyValues map[string]string) {
	for err != nil {
		if e, ok := err.(errorglue.ErrorHasData); ok {
			key, value := e.KeyValue()
			if key == "" { // for the slice
				list = append([]string{value}, list...)
			} else { // for the map
				if keyValues == nil {
					keyValues = map[string]string{key: value}
				} else if _, ok := keyValues[key]; !ok {
					keyValues[key] = value
				}
			}
		}
		err = errors.Unwrap(err)
	}
	return
}

// error116.ErrorList returns all error instances from a possible error chain.
// If err is nil an empty slice is returned.
// If err does not have associated errors, a slice of err is returned.
// otherwise, the first error of the returned slice is err followed by
// other errors oldest first.
// Cyclic error values are dropped
func ErrorList(err error) (errs []error) {
	if err == nil {
		return
	}
	err0 := err
	errMap := map[error]bool{err: true}
	for err != nil {
		if e, ok := err.(errorglue.RelatedError); ok {
			if e2 := e.AssociatedError(); e2 != nil {
				if _, ok := errMap[e2]; !ok {
					errs = append([]error{e2}, errs...)
					errMap[e2] = true
				}
			}
		}
		err = errors.Unwrap(err)
	}
	return append([]error{err0}, errs...)
}

// error116.HasStack detects if the error chain already contains a stack trace
func HasStack(err error) (hasStack bool) {
	var e errorglue.ErrorCallStacker
	return errors.As(err, &e)
}

// error116.IsWarning determines if an error has been flagged as a warning.
// error116.Warning() flags an error to be of warning level
func IsWarning(err error) bool {
	var warning *errorglue.WarningType
	return errors.Is(err, warning)
}

// error116.PackFunc returns the package name and function name
// of the caller:
//   error116.FuncName
func PackFunc() (packageDotFunction string) {
	return errorglue.NewCodeLocation(1).PackFunc()
}

// error116.Short gets a one-line location string similar to printf %-v and ShortFormat.
// Short() does not print stack traces, data and associated errors.
// Short() does print a one-liner of the error message and a brief code location:
//   error-message at error116.(*csTypeName).FuncName-chainstring_test.go:26
func Short(err error) string {
	return errorglue.ChainString(err, errorglue.ShortFormat)
}

// error116.Long() gets a comprehensive string representation similar to printf %+v and LongFormat.
// ShortFormat does not print stack traces, data and associated errors.
// Long() prints full stack traces, string key-value and list values for both the error chain
// of err, and associated errors and their chains
//   error-message
//     github.com/haraldrudell/parl/error116.(*csTypeName).FuncName
//       /opt/sw/privates/parl/error116/chainstring_test.go:26
//     runtime.goexit
//       /opt/homebrew/Cellar/go/1.17.8/libexec/src/runtime/asm_arm64.s:1133
func Long(err error) string {
	return errorglue.ChainString(err, errorglue.LongFormat)
}
