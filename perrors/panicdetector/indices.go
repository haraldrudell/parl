/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package panicdetector

import (
	"strings"

	"github.com/haraldrudell/parl/pruntime"
)

const (
	runtimePrefix = "runtime."
)

// Indices examines a stack to see if it includes a panic. Thread-safe
//   - isPanic is true if the stack includes a panic
//   - stack[recoveryIndex] is the code line of the deferred function containing recovery() invocation
//   - stack[panicIndex] is the code line causing the panic located in
//     the function deferring or any code that function invokes
//   - for panic detection to work, the code line of defer() invocation
//     must be included in stack
//   - note: the location in the deferring function is returned as the last
//     line of the deferring function
//   - [pruntime.Stack] is the parsed output of [runtime.Stack]
func Indices(stack pruntime.Stack) (isPanic bool, recoveryIndex, panicIndex int) {

	// scan for newest frame for the Go runtime’s defer invoker
	//	- [runtime.Stack], unlike [runtime.Callers], hides runtime frames handling the panic,
	//		so only look for panic()
	var found = 0
	var frames = stack.Frames()
	var stackLength = len(frames)
	var pd = panicDetectorOne
	for i := 0; i < stackLength; i++ {
		var funcName = frames[i].Loc().FuncName
		if i > 0 && funcName == pd.runtimeDeferInvokerLocation {
			recoveryIndex = i - 1
			found++
			if found == 2 {
				break
			}
		}
		if i+1 < stackLength && funcName == pd.runtimePanicFunctionLocation {

			// scan for end of runtime functions
			for panicIndex = i + 1; panicIndex+1 < stackLength; panicIndex++ {
				if !strings.HasPrefix(frames[panicIndex].Loc().FuncLine(), runtimePrefix) {
					break // this frame not part of the runtime
				}
			}

			found++
			if found == 2 {
				break
			}
		}
	}
	isPanic = found == 2

	return
}
