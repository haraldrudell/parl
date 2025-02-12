/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"fmt"
)

// Stack is implemented by:
//   - [github.com/haraldrudell/parl.stack]
//   - [github.com/haraldrudell/parl/pdebug.stack]
type Stack interface {
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
	GoFunction() (goFunction *CodeLocation)
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
	fmt.Stringer
}
