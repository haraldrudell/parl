/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

/*
Package pruntime provides an interface to the Go standard library’s runtime package using
only serializable simple types.

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
