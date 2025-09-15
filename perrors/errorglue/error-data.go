/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"fmt"
)

// errorData contains single key-value pair strings, both that may be empty
//   - rich error: has ChainString method
//   - [github.com/haraldrudell/parl/perrors.ErrorData] aggregates:
//   - — values with empty key into a slice
//   - — values with non-empty key into a map, overwriting older values
//   - [github.com/haraldrudell/parl/perrors/errorglue.ChainString] with formats:
//   - —  [github.com/haraldrudell/parl/perrors/errorglue.LongFormat] or
//   - — [github.com/haraldrudell/parl/perrors/errorglue.LongSuffix]
//   - — prints key:value with each error
type errorData struct {
	RichError
	key   string
	value string
}

// errorData behaves like an error
var _ error = &errorData{}

// errorData implements the error116.ErrorHasData interface
var _ ErrorHasData = &errorData{}

// errorData has an error chain
var _ Unwrapper = &errorData{}

// errorData can be used with fmt.Printf %v
var _ fmt.Formatter = &errorData{}

// errorData can be used by error116.ChainStringer()
var _ ChainStringer = &errorData{}

// NewErrorData returns an error containg a key/value pair
func NewErrorData(err error, key, value string) (e2 error) {
	return &errorData{*newRichError(err), key, value}
}

// KeyValue returns the assoiacted key/value pair
func (e *errorData) KeyValue() (key, value string) {
	if e == nil {
		return
	}
	return e.key, e.value
}

// ChainString returns a string representation
//   - invoked by [github.com/haraldrudell/parl/perrors/errorglue.ChainString]
func (e *errorData) ChainString(format CSFormat) (s string) {

	// nil case
	if e == nil {
		return
	}

	switch format {
	case ShortSuffix:
		return
	case DefaultFormat, ShortFormat: // no data
		s = e.Error()
		return
	case LongSuffix:
	case LongFormat:
		s = fmt.Sprintf("%s [%T]", e.Error(), e)
	default:
		return
	}
	// LongSuffix or LongFormat

	// build “key: value”
	//	- may be empty
	var keyValue = e.value
	if e.key != "" {
		keyValue = e.key + ":\x20" + keyValue
	}

	// append to s
	if s == "" {
		s = keyValue
	} else if keyValue != "" {
		s = s + "\n" + keyValue
	}

	return
}
