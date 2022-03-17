/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

/*
Package error116 adds stack traces and maps of values to error values.

fmt.Printf("%+v", err) stack, long lines

fmt.Printf("%-v", err) stack, short lines

fmt.Printf("%s", err)

fmt.Printf("%q", err)

© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
*/
package error116

import (
	"errors"
	"fmt"
	"strings"
)

// ErrorCallStacker indicates that this error is a call stack wrapping another error
type ErrorCallStacker interface {
	StackTrace() StackSlice
}

type Errors interface {
	error // Error() string: the short error message
	Long() string
}

// error116.Wrapper is an interface indicating error-chain capabilities.
// It is not public in errors package
type Wrapper interface {
	Unwrap() error // Unwrap returns the next error in the chain or nil
}

// ErrorHasData is an interface providing the ErrorHasData map
type ErrorHasData interface {
	error
	GetMap() (ed DataMap)
}

// ErrorHasCode allows an error to classify itself
type ErrorHasCode interface {
	error
	ErrorCode(code string) (hasCode bool)
	ErrorCodes(codes []string) (has []string)
}

type ErrorHasList interface {
	error
	ErrorList() (errors []error)
	Append(extra error) (err error)
	Count() int
}

// DataMap is a map of values associated with an error
type DataMap map[string]string

// HasStack detects if the error chain already contains a stack trace
func HasStack(err error) (hasStack bool) {
	var e ErrorCallStacker
	return errors.As(err, &e)
}

// GetStackTrace gets the last stack trace
func GetStackTrace(err error) (stack StackSlice) {
	var e ErrorCallStacker
	if errors.As(err, &e) {
		stack = e.StackTrace()
	}
	return
}

// GetInnerMostStack gets the oldest stack trace in the error chain
func GetInnerMostStack(err error) (stack StackSlice) {
	var e ErrorCallStacker
	for errors.As(err, &e) {
	}
	if e != nil {
		stack = e.StackTrace()
	}
	return
}

// DumpChain gets a space-separated string of error implementation type names
func DumpChain(err error) (s string) {
	var strs []string
	for err != nil {
		strs = append(strs, fmt.Sprintf("%T", err))
		err = errors.Unwrap(err)
	}
	s = strings.Join(strs, "\x20")
	return
}

func Errp(errp *error) func(e error) {
	if errp == nil {
		panic("Errp with nil argument")
	}
	return func(e error) {
		*errp = AppendError(*errp, e)
	}
}

// GetAllMaps retrieves all maps in the error chain, last at index 0
func GetAllMaps(err error) (maps []DataMap) {
	var e ErrorHasData
	for errors.As(err, &e) {
		maps = append(maps, e.GetMap())
	}
	return
}

// GetLastMap retrieves the last map in the error chain
func GetLastMap(err error) (m DataMap) {
	var e ErrorHasData
	errors.As(err, &e)
	if e != nil {
		m = e.GetMap()
	}
	return
}
