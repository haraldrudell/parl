/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"fmt"
	"io"
	"strconv"
)

// RichError is an error chain that behaves like fmt.Formatter.
// this allows for custom print-outs using %+v and %-v
// RichError has publics Error() Unwrap() Format()
type RichError struct {
	ErrorChain
}

var _ error = &RichError{}         // RichError behaves like an error
var _ Wrapper = &RichError{}       // RichError is an error chain
var _ fmt.Formatter = &RichError{} // RichError has features for fmt.Printf

func newRichError(err error) *RichError {
	return &RichError{*newErrorChain(err)}
}

// Format provides the fmt.Formatter function
func (e *RichError) Format(s fmt.State, verb rune) {
	if IsValueVerb(verb) {
		if format := PrintfFormat(s); format != DefaultFormat {
			ChainString(e, format)
		}
		io.WriteString(s, e.Error())
	} else if IsStringVerb(verb) {
		io.WriteString(s, e.Error())
	} else if IsQuoteVerb(verb) {
		io.WriteString(s, strconv.Quote(e.Error()))
	}
}

// IsPlusFlag determines if fmt.State has the '+' flag
func IsPlusFlag(s fmt.State) bool {
	return s.Flag('+')
}

// IsMinusFlag determines if fmt.State has the '-' flag
func IsMinusFlag(s fmt.State) bool {
	return s.Flag('-')
}

// IsValueVerb determines if the rune corresponds to the %v value verb
func IsValueVerb(r rune) bool {
	return r == 'v'
}

// IsStringVerb determines if the rune corresponds to the %s string verb
func IsStringVerb(r rune) bool {
	return r == 's'
}

// IsQuoteVerb determines if the rune corresponds to the %q quote verb
func IsQuoteVerb(r rune) bool {
	return r == 'q'
}
