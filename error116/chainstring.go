/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

import (
	"errors"
	"fmt"
)

const (
	// DefaultFormat %v output
	DefaultFormat Format = iota + 1
	// ShortFormat one-line location %-v
	ShortFormat
	// LongFormat full stack trace %+v
	LongFormat
	// ShortSuffix one-line without message
	ShortSuffix
	// LongSuffix full stack trace without message
	LongSuffix
)

// ChainStringer converts an error chain to string format
type ChainStringer interface {
	ChainString(format Format) string
}

// Format describes string conversion of an error chain
type Format byte

// Short stringifies errors with one-line extra data
func Short(err error) string {
	return ChainString(err, ShortFormat)
}

// Long stringifies errors with all extra data
func Long(err error) string {
	return ChainString(err, LongFormat)
}

// ChainString  can print an error chain of any errors
func ChainString(err error, format Format) (s string) {
	var didError bool
	for err != nil {
		if chainStringer, _ := err.(ChainStringer); chainStringer != nil {
			s += chainStringer.ChainString(SuffixFormat(format, didError))
			break // error.ChainString prints everything properly, we are done
		}
		if !didError {
			s += err.Error() // prints a chain of wrapped errors, but with no extra data
			if format == DefaultFormat {
				break // default format calls or nothing else
			}
			didError = true // string error messages have already been printed
		}
		err = errors.Unwrap(err)
	}
	return
}

// SuffixFormat takes into account if the wrapped error chain has already been printed
func SuffixFormat(format Format, didError bool) (f2 Format) {
	if didError && format == LongFormat {
		f2 = LongSuffix
	} else if didError && format == ShortFormat {
		f2 = ShortSuffix
	} else {
		f2 = format
	}
	return
}

// PrintfFormat gets Format for the Printf v verb
func PrintfFormat(s fmt.State) Format {
	if s.Flag('+') {
		return LongFormat
	} else if s.Flag('-') {
		return ShortFormat
	}
	return DefaultFormat
}
