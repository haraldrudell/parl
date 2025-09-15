/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"fmt"

	"github.com/haraldrudell/parl/pruntime"
)

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
func ErrorfPF(format string, a ...any) (err error) {
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

const (
	e116StacknFramesToSkip = 1
	// stack frames to skip in [New] [Errorf]
	e116StackFrames = 1
	// stack frames to skip in [NewPF]
	perrNewFrames = 1
)
