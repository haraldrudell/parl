/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

// PackFunc returns the package name and function name
// of the caller:
//
// [FuncName]
func PackFunc(skipFrames ...int) (packageDotFunction string) {

	// get skip
	var skip int
	if len(skipFrames) > 0 {
		skip = skipFrames[0]
	}
	if skip < 0 {
		skip = 0
	}

	var cL = NewCodeLocation(e116PackFuncStackFrames + skip)
	packageDotFunction = cL.Name()
	if pack := cL.Package(); pack != "main" {
		packageDotFunction = pack + "." + packageDotFunction
	}
	return
}

const (
	// stack frames in [PackFunc]
	e116PackFuncStackFrames = 1
)
