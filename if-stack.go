/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/pruntime"

// Stack contains a stack trace parsed into basic type only datapoints
//   - stack trace from [runtime.Stack] or [debug.Stack]
type Stack interface {
	// thread ID 1… for the thread requesting the stack trace
	//	- ThreadID is comparable and has IsValid and String methods
	//	- ThreadID is typically an incremented 64-bit integer with
	//		main thread having ID 1
	ID() ThreadID
	// a word indicating thread status, typically word “running”
	Status() ThreadStatus
	// true if the thread is the main thread
	//	- false for a launched goroutine
	IsMain() (isMain bool)
	// A list of code locations for this thread
	//	- index [0] is the most recent code location, typically the invoker requesting the stack trace
	//	- includes invocation argument values
	Frames() (frames []Frame)
	// the goroutine function used to launch this thread
	//	- if IsMain is true, zero-value. Check using GoFunction().IsSet()
	//	- never nil
	GoFunction() (goFunction *pruntime.CodeLocation)
	// the code location of the go statement creating this thread
	//	- if IsMain is true, zero-value. Check with Creator().IsSet()
	//	- never nil
	Creator() (creator *pruntime.CodeLocation)
	// Shorts lists short code locations for all stack frames, most recent first:
	// Shorts("prepend") →
	//  prepend Thread ID: 1
	//  prepend main.someFunction()-pruntime.go:84
	//  prepend main.main()-pruntime.go:52
	Shorts(prepend string) (s string)
	// String is a multi-line stack trace, most recent code location first:
	//  ID: 18 IsMain: false status: running␤
	//  main.someFunction({0x100dd2616, 0x19})␤
	//  ␠␠pruntime.go:64␤
	//  cre: main.main-pruntime.go:53␤
	String() (s string)
}

// Frame represents an executing code location, ie. a code line in source code
//   - parl.Frame is similar to [runtime.Frame] returned by [runtime.CallersFrames]
//     but has only basic types, ie. it can be printed, stored and transferred
type Frame interface {
	// the code location for this frame, never nil
	Loc() (location *pruntime.CodeLocation)
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
