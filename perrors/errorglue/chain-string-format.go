/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import "strconv"

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

// CSFormat describes string conversion of an error chain
type CSFormat uint8

func (csFormat CSFormat) String() (s string) {
	var ok bool
	if s, ok = csFormatMap[csFormat]; !ok {
		s = "?" + strconv.Itoa(int(csFormat))
	}
	return
}

var csFormatMap = map[CSFormat]string{
	DefaultFormat: "DefaultFormat",
	ShortFormat:   "ShortFormat",
	LongFormat:    "LongFormat",
	ShortSuffix:   "ShortSuffix",
	LongSuffix:    "LongSuffix",
}
