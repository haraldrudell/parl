/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import "strconv"

const (
	// DefaultFormat is like printf %v, printf %s and error.Error() “message”
	//	- data, stack traces and associated errors are not printed
	//	- code location is not printed
	DefaultFormat CSFormat = iota + 1
	// CodeLocation DefaultFormat with location added “message at runtime/panic.go:914”
	//	- location is:
	//	- — oldest panic code-line in any stack trace of err or its error chain
	//	- — code-line creating the oldest error in err and its chain that has
	//		a stack trace, when none of the errors hold a panic
	CodeLocation
	// ShortFormat has one-line location similar to printf %-v
	//	- data, and stack traces are not printed
	//	- associated errors are printed
	//	- if the error or its chain contains a stack trace, a code location is output
	//	- if any stack trace contains a panic, that oldest panic location is printed
	//	- “error-message at error116.(*csTypeName).FuncName-chainstring_test.go:26”
	ShortFormat
	// LongFormat is similar to printf %+v.
	//	- prints data, stack traces and associated errors
	//	- if any stack trace contains a panic, that oldest panic location is printed after message
	//
	// output:
	//	error-message
	//	  github.com/haraldrudell/parl/error116.(*csTypeName).FuncName
	//	    /opt/sw/privates/parl/error116/chainstring_test.go:26
	//	  runtime.goexit
	//	    /opt/homebrew/Cellar/go/1.17.8/libexec/src/runtime/asm_arm64.s:1133
	LongFormat
	// ShortSuffix is code-location “runtime/panic.go:914”
	//	- no leading “ at ”
	//	- if no stack is present, empty string
	ShortSuffix
	// LongSuffix full stack trace without message
	LongSuffix
)

// CSFormat describes string conversion of an error chain
//   - DefaultFormat ShortFormat LongFormat ShortSuffix LongSuffix
type CSFormat uint8

func (csFormat CSFormat) String() (s string) {
	var ok bool
	if s, ok = csFormatMap[csFormat]; !ok {
		s = "?" + strconv.Itoa(int(csFormat))
	}
	return
}

// map for quick printable-string lookup
var csFormatMap = map[CSFormat]string{
	DefaultFormat: "DefaultFormat",
	ShortFormat:   "ShortFormat",
	LongFormat:    "LongFormat",
	ShortSuffix:   "ShortSuffix",
	LongSuffix:    "LongSuffix",
}
