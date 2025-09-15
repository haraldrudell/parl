/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"fmt"
	"io"
	"strconv"

	"github.com/haraldrudell/parl/pfmt"
)

// RichError is an error chain that behaves like fmt.Formatter.
// this allows for custom print-outs using %+v and %-v
// RichError has publics Error() Unwrap() Format()
type RichError struct {
	ErrorChain
}

// RichError behaves like an error
var _ error = &RichError{}

// RichError is an error chain
var _ Unwrapper = &RichError{}

// RichError has features for fmt.Printf
var _ fmt.Formatter = &RichError{}

func newRichError(err error) *RichError {
	return &RichError{*newErrorChain(err)}
}

// Format provides the fmt.Formatter function
func (e *RichError) Format(s fmt.State, verb rune) {
	if pfmt.IsValueVerb(verb) {
		if format := PrintfFormat(s); format != DefaultFormat {
			ChainString(e, format)
		}
		io.WriteString(s, e.Error())
	} else if pfmt.IsStringVerb(verb) {
		io.WriteString(s, e.Error())
	} else if pfmt.IsQuoteVerb(verb) {
		io.WriteString(s, strconv.Quote(e.Error()))
	}
}
