/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"fmt"

	"github.com/haraldrudell/parl/pruntime"
)

const (
	// stack frames to skip [Is] [IsPF] [Is2] [Is2PF]
	isStackFrames = 1
)

// Is returns true if *errp contains a non-nil error
//   - if return value is true and format is not empty string, *errp is updated with
//     fmt.Errorf using format and a, typically including “%w” and an error
//   - if *errp is non-nil and does not have a stack, a stack is inserted into
//     its error chain
//   - errp cannot be nil or panic
func Is(errp *error, format string, a ...interface{}) (isBad bool) {
	if errp == nil {
		panic(NewPF("errp nil"))
	}
	err := *errp
	if err == nil {
		return // no error exit
	}
	if format != "" {
		err = fmt.Errorf(format, a...)
	}
	if !HasStack(err) {
		err = Stackn(err, isStackFrames)
	}
	*errp = err
	return true
}

// IsPF returns true if *errp contains a non-nil error
//   - package and function identifiers are prepended
//   - if return value is true and format is not empty string, *errp is updated with
//     fmt.Errorf using format and a, typically including “%w” and an error
//   - if *errp is non-nil and does not have a stack, a stack is inserted into
//     its error chain
//   - errp cannot be nil or panic
func IsPF(errp *error, format string, a ...interface{}) (isBad bool) {
	if errp == nil {
		panic(NewPF("errp nil"))
	}
	err := *errp
	if err == nil {
		return false // no error exit
	}
	var PF = pruntime.PackFunc(isStackFrames)
	if format == "" {
		err = fmt.Errorf("%s %w", PF, err)
	} else {
		err = fmt.Errorf("%s "+format, append([]interface{}{PF}, a...)...)
	}
	if !HasStack(err) {
		err = Stackn(err, isStackFrames)
	}
	*errp = err
	return true
}
