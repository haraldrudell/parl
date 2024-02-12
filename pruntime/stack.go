/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

/*
Package pruntime provides an interface to the Go standard library’s runtime package using
only serializable simple types

Stack traces and code locations have several formats:

	codeLocation := pruntime.NewCodeLocation(0)
	codeLocation.Base() // package and type
	  → mypackage.(*MyType).MyFunc
	codeLocation.PackFunc() // very brief
	  → mypackage.MyFunc
	codeLocation.Name(): // function name only
	  → MyFunc
	codeLocation.Short() // line, no package path
	  → mypackage.(*MyType).MyFunc-myfile.go:19
	codeLocation.Long() // uniquely identifiable
	  → codeberg.org/haraldrudell/mypackage.(*MyType).MyFunc-myfile.go:19
	codeLocation.Full() // everything
	  → codeberg.org/haraldrudell/mypackage.(*MyType).MyFunc-/fs/mypackage/myfile.go:19
	codeLocation.String() // two lines
	  → "codeberg.org/haraldrudell/mypackage.(*MyType).MyFunc\n  /fs/mypackage/myfile.go:19"

Stack can determine where a goroutine was created and whether this is the main thread

	pruntime.GoRoutineID()  → 1
	pruntime.NewStack(0).Creator.Short()  → main.main-pruntime.go:30
	fmt.Println(pruntime.NewStack(0).IsMainThread)  → true
	pruntime.NewStack(0).Frames[0].Args  → (0x104c12c60?)
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
