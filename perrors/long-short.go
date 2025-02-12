/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import "github.com/haraldrudell/parl/perrors/errorglue"

// error116.Long() gets a comprehensive string representation similar to printf %+v and LongFormat.
// ShortFormat does not print stack traces, data and associated errors.
// Long() prints full stack traces, string key-value and list values for both the error chain
// of err, and associated errors and their chains
//
//	error-message
//	  github.com/haraldrudell/parl/error116.(*csTypeName).FuncName
//	    /opt/sw/privates/parl/error116/chainstring_test.go:26
//	  runtime.goexit
//	    /opt/homebrew/Cellar/go/1.17.8/libexec/src/runtime/asm_arm64.s:1133
func Long(err error) (s string) {
	return errorglue.ChainString(err, errorglue.LongFormat)
}

// LongShort picks output format
//   - no error: “OK”
//   - no panic: “error-message at runtime.gopanic:26”
//   - panic: long format with all stack traces and values
//   - associated errors are always printed
func LongShort(err error) (message string) {
	var format errorglue.CSFormat
	if err == nil {
		format = errorglue.ShortFormat
	} else if isPanic, _, _, _ := IsPanic(err); isPanic {
		format = errorglue.LongFormat
	} else {
		format = errorglue.ShortFormat
	}
	message = errorglue.ChainString(err, format)

	return
}

// perrors.Short gets a one-line location string similar to printf %-v and ShortFormat.
// Short() does not print stack traces, data and associated errors.
// Short() does print a one-liner of the error message and a brief code location:
//
//	error-message at error116.(*csTypeName).FuncName-chainstring_test.go:26
func Short(err error) (s string) {
	return errorglue.ChainString(err, errorglue.ShortFormat)
}
