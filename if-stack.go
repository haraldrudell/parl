/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/pruntime"

type Stack interface {
	ID() ThreadID          // the thread ID 1… for the thread requesting the stack trace. main thread has ID 1.
	Status() ThreadStatus  // a word indicating thread status, typically word “running”
	IsMain() (isMain bool) // true if the thread is the main thread
	// A list of code locations for this thread.
	// [0] is the most recent code location, typically the invoker requesting the stack trace.
	Frames() (frames []Frame)
	// goFunction is the function that a goroutine launched
	//	- is isMain is true, it is the zero-value
	//	- never nil
	GoFunction() (goFunction *pruntime.CodeLocation)
	// the code location of the go statement creating this thread.
	//	- if IsMain true, this field is zero-value, check with CodeLocation.IsSet
	//	- never nil
	Creator() (creator *pruntime.CodeLocation)
	// Shorts lists short code locations for all stack frames, most recent first:
	// Shorts("prepend") →
	//  prepend Thread ID: 1
	//  prepend main.someFunction()-pruntime.go:84
	//  prepend main.main()-pruntime.go:52
	Shorts(prepend string) (s string)
	// String is a multi-line stack trace, most recent code location first:
	//  ID: 18 IsMain: false status: running
	//  main.someFunction({0x100dd2616, 0x19})
	//  pruntime.go:64
	//  cre: main.main-pruntime.go:53
	String() (s string)
}

type Frame interface {
	Loc() (location *pruntime.CodeLocation) // the code location for this frame, never nil
	Args() (args string)                    // invocation values like "(0x14000113040)"
	// package funtion name, base filename ad line number
	String() (s string)
}
