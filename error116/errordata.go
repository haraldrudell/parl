/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

import (
	"errors"
	"fmt"
	"io"
)

type errorData struct {
	ErrorChain
	DataMap
}

// ErrorData adds a datamap to an error chain
func ErrorData(err error, m DataMap) (e2 error) {
	if err == nil {
		e2 = err
		return
	}
	if m == nil {
		m = DataMap{}
	}
	return &errorData{ErrorChain{err}, m}
}

func (e *errorData) GetMap() (m DataMap) {
	m = e.DataMap
	return
}

func (e *errorData) mapString() string {
	s := ""
	if m := e.DataMap; m != nil {
		for key, value := range m {
			/*
				if t, ok := value.(time.Time); ok {
					value = t.Format(time.RFC3339Nano)
				}
			*/
			s += fmt.Sprintf("\n- %s: %v", key, value)
		}
	}
	return s
}

func (e *errorData) ChainString(format Format) (s string) {
	printMessage := format == DefaultFormat || format != ShortSuffix && format != LongSuffix
	if e == nil {
		return
	} else if printMessage {
		if s = e.Error(); format == DefaultFormat {
			return
		}
	}

	s += e.mapString()

	var chainString ChainStringer
	if errors.As(e.error, &chainString); chainString == nil {
		return
	}
	if format == LongFormat {
		format = LongSuffix
	} else if format == ShortFormat {
		format = ShortSuffix
	}
	s += chainString.ChainString(format)

	return
}

// Format provides the fmt.Formatter function
func (e *errorData) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if format := PrintfFormat(s); format != DefaultFormat {
			io.WriteString(s, e.ChainString(format))
			return
		}
		fallthrough // %v is same as %s
	case 's':
		io.WriteString(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	}
}
