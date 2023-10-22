/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"runtime"
)

// GetCodeLocation converts a runtime stack frame to a code location
// stack frame.
// runtime contains opaque types while code location consists of basic
// types int and string only
func GetCodeLocation(rFrame *runtime.Frame) (cl *CodeLocation) {
	// runtime.Frame:
	// PC uintptr, Func *Func, Function string, File string, Line int, Entry uintptr, funcInfo funcInfo
	return &CodeLocation{
		File:     rFrame.File,
		Line:     rFrame.Line,
		FuncName: rFrame.Function,
	}
}
