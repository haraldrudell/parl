/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import "github.com/haraldrudell/parl/perrors/errorglue"

// perrors.Short gets a one-line location string similar to printf %-v and ShortFormat.
// Short() does not print stack traces, data and associated errors.
// Short() does print a one-liner of the error message and a brief code location:
//
//	error-message at error116.(*csTypeName).FuncName-chainstring_test.go:26
func Short(err error) (s string) {
	return errorglue.ChainString(err, errorglue.ShortFormat)
}
