/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"strconv"
)

// Frame represents an executing code location, ie. a code line in source code
//   - parl.Frame is similar to [runtime.Frame] returned by [runtime.CallersFrames]
//     but has only basic types, ie. can be stored or transferred
//   - [pdebug.Frame] implements [parl.Frame]
//   - Frame is a value container only created by [pdebug.NewStack].
//     Frame extends [pruntime.CodeLocation] with argument values
type FrameR struct {
	CodeLocation
	// function argument values like "(1, 0x14000113040)"
	//	- values of value-types like int are displayed
	//	- most types appear as a pointer value “0x…”
	args string
}

var _ Frame = &FrameR{}

// the code location for this frame, never nil
func (f *FrameR) Loc() (location *CodeLocation) { return &f.CodeLocation }

// function argument values like "(1, 0x14000113040)"
//   - values of basic types like int are displayed
//   - most types appear as a pointer value “0x…”
func (f *FrameR) Args() (args string) { return f.args }

// prints the Frame suitable to be part of a stack trace
//   - fully qualified package name with function or type and method
//     and argument values
//   - absolute path to source file and line number
//
// output:
//
//	github.com/haraldrudell/parl/pdebug.TestFrame(0x1400014a340)␤
//	␠␠frame_test.go:15
func (f *FrameR) String() (s string) {
	return f.CodeLocation.FuncName + f.args + "\n" +
		"\x20\x20" + f.CodeLocation.File + ":" + strconv.Itoa(f.CodeLocation.Line)
}
