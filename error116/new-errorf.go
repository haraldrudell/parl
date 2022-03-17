/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

import (
	"errors"
	"fmt"
)

// New does errors.New and error116.Stack
func New(s string) error {
	if s == "" {
		s = "StackNew from " + NewCodeLocation(1).Short()
	}
	return Stack(errors.New(s))
}

// Errorf wraps error and ensure a stack exists
func Errorf(format string, a ...interface{}) (err error) {
	err = fmt.Errorf(format, a...)
	if HasStack(err) {
		return
	}
	return Stackn(err, 1)
}

// Stack ensures the error chain has a stack trace
func Stack(err error) (err2 error) {
	if err == nil || HasStack(err) {
		err2 = err
		return
	}
	err2 = &errorStack{ErrorChain{err}, NewStackSlice(stackFramesToSkip)}
	return
}

// Stackn is Stack but allows to skip caller stack frames and will always add a stack
func Stackn(err error, framesToSkip int) (err2 error) {
	if err == nil {
		return
	}
	if framesToSkip < 0 {
		framesToSkip = 0
	}
	err2 = &errorStack{ErrorChain{err}, NewStackSlice(stackFramesToSkip + framesToSkip)}
	return
}
