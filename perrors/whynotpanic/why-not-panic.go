/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package whynotpanic

import (
	"fmt"
	"strings"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors/errorglue"
	"github.com/haraldrudell/parl/perrors/panicdetector"
)

var whyTemplate = strings.Join([]string{
	" WHY_NOT_PANIC: isPanic: %t recoveryIndex: %d panicIndex: %d error-nil: %t stacks: %d",
	"WhyNotPanic describes why a panic was not detected for an error value",
	"— isPanic: false means the panic detector did not find a stack trace containing a Go panic",
	"— if recoveryIndex or panicIndex is 0, it means that the line was not found in any stack trace",
	"— error-nil: true means the error value was nil and did not contain an error",
	"— stacks is the number of stack traces found in the errors’ error chain",
	"— if stacks is 0, it means the error is not the result of any of parl’s recovery functions",
	"— panicIndex is the stack trace line before a Go detected a panic. It is the code line causing panic",
	"— the known runtime package reference invoked on Go panic: ‘%s’",
	"— recoveryIndex is the stack frame after Go panic handling concludes",
	"— the known runtime package reference invoking deferred functions after a panic is: ‘%s’",
	"error message for last error with stack trace: %s",
	"last stack trace: %s",
	"— end WHY_NOT_PANIC",
}, "\n")

// WhyNotPanic returns a printble string explaining panic-data on err
func WhyNotPanic(err error) (s string) {

	// find any panic stack trace in the err error chain
	var isPanic, stack, recoveryIndex, panicIndex, numberOfStacks, errorWithStack = errorglue.FirstPanicStack(err)
	if isPanic {
		return // panic is detected return: empty string
	}

	var message string
	if errorWithStack != nil {
		message = "‘" + errorWithStack.Error() + "’"
	} else {
		message = "none"
	}
	var lastStack string
	if st, ok := stack.(parl.Stack); ok {
		if len(st.Frames()) > 0 {
			lastStack = stack.String()
		}
	}
	if lastStack == "" {
		lastStack = "none"
	}

	var deferS, panicS = panicdetector.PanicDetectorValues()

	s = fmt.Sprintf(whyTemplate,
		isPanic, recoveryIndex, panicIndex, err == nil, numberOfStacks,
		deferS,
		panicS,
		message, lastStack,
	)
	return
}
