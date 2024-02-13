/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

// Frame represents an executing code location, ie. a code line in source code
//   - parl.Frame is similar to [runtime.Frame] returned by [runtime.CallersFrames]
//     but has only basic types, ie. it can be printed, stored and transferred
//   - a lis of Frame is returned by [Stack.Frames]
type Frame interface {
	// the code location for this frame, never nil
	Loc() (location *CodeLocation)
	// function argument values like “(1, 0x14000113040)”
	//	- values of basic types like int are displayed
	//	- most types appear as a pointer value “0x…”
	Args() (args string)
	// prints the Frame suitable to be part of a stack trace
	//   - fully qualified package name with function or type and method
	//     and argument values
	//   - absolute path to source file and line number
	//
	// output:
	//
	//	github.com/haraldrudell/parl/pdebug.TestFrame(0x1400014a340)␤
	//	␠␠frame_test.go:15
	String() (s string)
}
