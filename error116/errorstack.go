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

const stackFramesToSkip = 2

type errorStack struct {
	ErrorChain
	StackSlice // slice
}

func (e *errorStack) StackTrace() (st StackSlice) {
	if e != nil {
		st = e.StackSlice
	}
	return
}

func (e *errorStack) ChainString(format Format) (s string) {
	if e == nil {
		return
	} else if format == DefaultFormat {
		s = e.Error()
		return
	}
	if format != ShortSuffix && format != LongSuffix {
		s = e.Error()
	}
	if format == LongFormat || format == LongSuffix {
		s += e.StackSlice.String()
	} else {
		s += e.StackSlice.Short()
	}
	var chainString ChainStringer
	errors.As(e.error, &chainString)
	if chainString == nil {
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

func (e *errorStack) Format(s fmt.State, verb rune) {
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
