/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

import (
	"errors"
	"fmt"

	"github.com/haraldrudell/parl/errorglue"
)

const (
	stackFramesToSkip = 1
	newFramesToSkip   = 1
)

// error116.New is similar to errors.New but ensures that the returned error
// has at least one stack trace associated
func New(s string) error {
	if s == "" { // ensure there is an error message
		s = "StackNew from " + errorglue.NewCodeLocation(1).Short()
	}
	return Stackn(errors.New(s), newFramesToSkip)
}

// error116.Errorf is similar to fmt.Errorf but ensures that the returned err
// has at least one stack trace associated
func Errorf(format string, a ...interface{}) (err error) {
	err = fmt.Errorf(format, a...)
	if HasStack(err) {
		return
	}
	return Stackn(err, 1)
}

// error116.Stack ensures the err has a stack trace
// associated.
// err can be nil in which nil is returned
func Stack(err error) (err2 error) {
	if err == nil || HasStack(err) {
		err2 = err
		return
	}
	err2 = errorglue.NewErrorStack(err, errorglue.NewStackSlice(stackFramesToSkip))
	return
}

// error116.Stackn always attaches a new stack trace to err and
// allows for skipping stack frames using framesToSkip.
// if err is nil, no action is taken
func Stackn(err error, framesToSkip int) (err2 error) {
	if err == nil {
		return
	}
	if framesToSkip < 0 {
		framesToSkip = 0
	}
	err2 = errorglue.NewErrorStack(err, errorglue.NewStackSlice(stackFramesToSkip+framesToSkip))
	return
}
