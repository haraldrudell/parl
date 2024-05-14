/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"fmt"
	"strings"

	"github.com/haraldrudell/parl/perrors/panicdetector"
	"github.com/haraldrudell/parl/pruntime"
	"golang.org/x/exp/slices"
)

func ListError(err error) (s string) {
	if err == nil {
		s = "ListError: OK"
		return
	}
	// the linked-list error chain needs to be traversed oldest error first
	//	- subsequent errors will report the same stack
	var errs []error
	for e := err; e != nil; e = errors.Unwrap(e) {
		errs = append(errs, e)
	}
	// errs now oldest first
	slices.Reverse(errs)
	// error texts, newest error first
	var sL = make([]string, len(errs))
	for i, e := range errs {
		if false {
			FirstPanicStack(err)
		}
		var stack pruntime.Stack
		var isPanicS string
		if errWithStack, ok := e.(ErrorCallStacker); ok {
			stack = errWithStack.StackTrace()
			if isPanic, recoveryIndex, panicIndex := panicdetector.Indices(stack); isPanic {
				isPanicS = " PANIC"
				_ = recoveryIndex
				_ = panicIndex
			}
		}
		var line = fmt.Sprintf(
			"ListError:%d(%d) %s%T “%s”",
			len(sL)-i, len(errs),
			isPanicS,
			e, e.Error(),
		)
		if stack != nil {
			line += "\n" + stack.String()
		}
		sL[len(sL)-i-1] = line
	}
	s = strings.Join(sL, "\n") + "\n— — —"
	return
}
