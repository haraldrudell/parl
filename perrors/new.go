/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"errors"

	"github.com/haraldrudell/parl/pruntime"
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
