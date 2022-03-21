/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"fmt"
)

type errorData struct {
	RichError
	key   string
	value string
}

var _ error = &errorData{}         // errorData behaves like an error
var _ ErrorHasData = &errorData{}  // errorData implements the error116.ErrorHasData interface
var _ Wrapper = &errorData{}       // errorData has an error chain
var _ fmt.Formatter = &errorData{} // errorData can be used with fmt.Printf %v
var _ ChainStringer = &errorData{} // errorData can be used by error116.ChainStringer()

func NewErrorData(err error, key, value string) (e2 error) {
	return &errorData{*NewRichError(err), key, value}
}

func (e *errorData) KeyValue() (key, value string) {
	if e == nil {
		return
	}
	return e.key, e.value
}

func (e *errorData) ChainString(format ErrorFormat) (s string) {
	if e == nil || format == ShortSuffix {
		return
	}
	if format != LongSuffix {
		s = e.Error()
		if format == DefaultFormat || format == ShortFormat {
			return
		}
	}

	// LongSuffix or LongFormat
	s2 := e.value
	if e.key != "" {
		s2 = e.key + ":\x20" + s2
	}
	if s == "" {
		s = s2
	} else if s2 != "" {
		s = s + "\n" + s2
	}

	return
}
