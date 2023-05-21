/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"errors"
	"fmt"

	"github.com/haraldrudell/parl/errorglue"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	e116StacknFramesToSkip = 1
	e116StackFrames        = 1
	perrNewFrames          = 1
)

// error116.New is similar to errors.New but ensures that the returned error
// has at least one stack trace associated
func New(s string) error {
	if s == "" { // ensure there is an error message
		s = "StackNew from " + pruntime.NewCodeLocation(perrNewFrames).Short()
	}
	return Stackn(errors.New(s), e116StackFrames)
}

func NewPF(s string) error {
	packFunc := pruntime.NewCodeLocation(perrNewFrames).PackFunc()
	if s == "" {
		s = packFunc
	} else {
		s = packFunc + "\x20" + s
	}
	return Stackn(errors.New(s), e116StackFrames)
}

// error116.Errorf is similar to fmt.Errorf but ensures that the returned err
// has at least one stack trace associated
func Errorf(format string, a ...interface{}) (err error) {
	err = fmt.Errorf(format, a...)
	if HasStack(err) {
		return
	}
	return Stackn(err, e116StackFrames)
}

func ErrorfPF(format string, a ...interface{}) (err error) {
	// format may include %w directives, meaning fmt.Errorf must be used
	// format may include numeric indices like %[1]s, meaning values cannot be prepended to a
	format = pruntime.NewCodeLocation(perrNewFrames).PackFunc() + "\x20" + format
	err = fmt.Errorf(format, a...)
	if HasStack(err) {
		return
	}
	return Stackn(err, e116StackFrames)
}

// error116.Stack ensures the err has a stack trace
// associated.
// err can be nil in which nil is returned
func Stack(err error) (err2 error) {
	if HasStack(err) {
		return err
	}
	return Stackn(err, e116StackFrames)
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
	err2 = errorglue.NewErrorStack(err, pruntime.NewStackSlice(e116StacknFramesToSkip+framesToSkip))
	return
}
