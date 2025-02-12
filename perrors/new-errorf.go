/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"errors"
	"fmt"

	"github.com/haraldrudell/parl/pruntime"
)

const (
	e116StacknFramesToSkip = 1
	// stack frames to skip in [New] [Errorf]
	e116StackFrames = 1
	// stack frames to skip in [NewPF]
	perrNewFrames = 1
)

// New is similar to [errors.New] but ensures that the returned error
// has at least one stack trace associated
//   - if s is empty “StackNew from …”
func New(s string) (err error) {

	// ensure there is an error message
	if s == "" {
		s = "StackNew from " + pruntime.NewCodeLocation(perrNewFrames).Short()
	}

	// add stack
	err = Stackn(errors.New(s), e116StackFrames)

	return
}

// NewPF is similar to [errors.New] but ensures that the returned error
// has at least one stack trace associated
//   - if s is empty “StackNew from …”
//   - prepends error message with package name and function identifiers
//   - “perrors NewPF s cannot be empty”
func NewPF(s string) (err error) {
	var packFunc = pruntime.NewCodeLocation(perrNewFrames).PackFunc()
	if s == "" {
		s = packFunc
	} else {
		s = packFunc + "\x20" + s
	}
	err = Stackn(errors.New(s), e116StackFrames)
	return
}

// Errorf is similar to [fmt.Errorf] but ensures that the returned err
// has at least one stack trace associated
func Errorf(format string, a ...interface{}) (err error) {
	err = fmt.Errorf(format, a...)
	if HasStack(err) {
		return
	}
	err = Stackn(err, e116StackFrames)
	return
}

// Errorf is similar to [fmt.Errorf] but ensures that the returned err
// has at least one stack trace associated
//   - prepends error message with package name and function identifiers
//   - “perrors NewPF s cannot be empty”
func ErrorfPF(format string, a ...interface{}) (err error) {
	// format may include %w directives, meaning fmt.Errorf must be used
	// format may include numeric indices like %[1]s, meaning values cannot be prepended to a
	format = pruntime.NewCodeLocation(perrNewFrames).PackFunc() + "\x20" + format
	err = fmt.Errorf(format, a...)
	if HasStack(err) {
		return
	}
	err = Stackn(err, e116StackFrames)
	return
}
