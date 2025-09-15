/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"testing"

	"github.com/haraldrudell/parl/pruntime"
)

func TestGetStackTrace(t *testing.T) {
	var errNoStack = errors.New("x")
	var errOneStack = NewErrorStack(errNoStack, pruntime.NewStack(0))
	var stack1 = errOneStack.(ErrorCallStacker).StackTrace()
	var errTwoStacks = NewErrorStack(errOneStack, pruntime.NewStack(1))
	var stack2 = errTwoStacks.(ErrorCallStacker).StackTrace()
	if len(stack1.Frames()) == len(stack2.Frames()) {
		t.Error("stacks same")
	}

	var stack pruntime.Stack

	// nil error should return nil
	stack = GetStackTrace(nil)
	if stack != nil {
		t.Errorf("stack not nil")
	}

	// no-stack error should return nil
	stack = GetStackTrace(errNoStack)
	if stack != nil {
		t.Errorf("stack not nil")
	}

	stack = GetStackTrace(errOneStack)
	if len(stack.Frames()) == 0 {
		t.Error("stack3 len 0")
	}
	if len(stack.Frames()) != len(stack1.Frames()) {
		t.Errorf("stack3 bad length %d exp %d", len(stack.Frames()), len(stack1.Frames()))
	}

	stack = GetStackTrace(errTwoStacks)
	if len(stack.Frames()) == 0 {
		t.Error("stack4 len 0")
	}
	if len(stack.Frames()) != len(stack2.Frames()) {
		t.Errorf("stack4 bad length %d exp %d", len(stack.Frames()), len(stack2.Frames()))
	}
}
