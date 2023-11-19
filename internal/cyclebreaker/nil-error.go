/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package cyclebreaker

import (
	"fmt"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	// count frame of [parl.NilError]
	nilErrorFrames = 1
)

// ErrNil is used with [errors.Is] to detect that a panic or error value was caused
// by a value that cannot be nil, such as a function argument to a new function,
// was nil
//   - —
//   - a nilValue type implements:
//   - — a dummy NilValueError method
//   - — an Is method returning true for errors implementing a NilValueError method
//
// Usage:
//
//	if errors.Is(err, parl.ErrNil) { …
var ErrNil = &nilValue{"value"}

// NilError returns an error used with panic() indicating that a value that cannot be nil,
// such as a function argument to a new function, was nil
//   - such panics typically indicate compile-time issues with code
//
// Usage:
//
//	func NewX(xValue *int) (x *X) {
//	  if xValue == nil {
//	    panic(parl.NilError("xValue")) // “somePackage.NewX xValue cannot be nil”
func NilError(valueName string) (err error) {
	var cL = pruntime.NewCodeLocation(nilErrorFrames)
	return perrors.Stackn(&nilValue{
		s: fmt.Sprintf("%s %s cannot be nil",
			cL.PackFunc(), // “somePackage.NewX”
			valueName,     // “xValue”
		)}, nilErrorFrames)
}

var _ nilValueIface = &nilValue{}

// nilValue is the type for NilError errors
type nilValue struct{ s string }

// nilValueIface looks for the NilValueError method
type nilValueIface interface{ NilValueError() }

// Error return the error message
//   - allows nilValue to implement the error interface
func (e *nilValue) Error() (message string) { return e.s }

// Is method allowing for nilValue values to detect other nilValues
func (e *nilValue) Is(err error) (is bool) {
	_, is = err.(nilValueIface)
	return
}

// dummy method NilValueError
func (e *nilValue) NilValueError() {}
