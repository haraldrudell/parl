/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"fmt"
	"strings"

	"github.com/haraldrudell/parl/pruntime"
)

const (
	runtimePrefix = "runtime."
)

// Indices examines a stack to see if it includes a panic. Thread-safe
//   - isPanic is true if the stack includes a panic
//   - stack[recoveryIndex] is the code line of the deferred function containing recovery invocation
//   - stack[panicIndex] is the code line causing the panic
func Indices(stack pruntime.StackSlice) (isPanic bool, recoveryIndex, panicIndex int) {
	found := 0
	length := len(stack)
	pd := panicDetectorOne
	for i := 0; i < length; i++ {
		funcName := stack[i].FuncName
		if i > 0 && funcName == pd.runtimeDeferInvokerLocation {
			recoveryIndex = i - 1
			found++
			if found == 2 {
				break
			}
		}
		if i+1 < length && funcName == pd.runtimePanicFunctionLocation {

			// scan for end of runtime functions
			for panicIndex = i + 1; panicIndex+1 < length; panicIndex++ {
				if !strings.HasPrefix(stack[panicIndex].FuncLine(), runtimePrefix) {
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

func WhyNotPanic(err error) (s string) {
	stack := GetInnerMostStack(err)
	isPanic, recoveryIndex, panicIndex := Indices(stack)
	if isPanic {
		return // panic is detected return: empty string
	}
	s = fmt.Sprintf(" WhyNotPanic: isPanic: %t recoveryIndex: %d panicIndex: %d\n"+
		"— recoveryIndex or panicIndex 0 means those lines were not found: unexpected\n"+
		"panicLine: the stack trace line before a panic() frame:\n"+
		"— the runtime location invoked by panic()\n%s\n"+
		"deferLine: the stack frame after a recover() frame:\n"+
		"— the runtime location invoking the recovering deferred function\n%s\n"+
		"error:%s\n\n",
		isPanic, recoveryIndex, panicIndex,
		panicDetectorOne.runtimePanicFunctionLocation,
		panicDetectorOne.runtimeDeferInvokerLocation,
		ChainString(err, LongFormat),
	)
	return
}
