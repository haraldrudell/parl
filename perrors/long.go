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
