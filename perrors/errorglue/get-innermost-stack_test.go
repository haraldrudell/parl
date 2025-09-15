/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue_test

import (
	"errors"
	"testing"

	"github.com/haraldrudell/parl/perrors/errorglue"
	"github.com/haraldrudell/parl/pruntime"
)

func TestGetInnerMostStack(t *testing.T) {
	var errNoStack = errors.New("x")
	var errOneStack = errorglue.NewErrorStack(errNoStack, pruntime.NewStack(0))
	var stack1 = errOneStack.(errorglue.ErrorCallStacker).StackTrace()
	var errTwoStacks = errorglue.NewErrorStack(errOneStack, pruntime.NewStack(1))
	var stack2 = errTwoStacks.(errorglue.ErrorCallStacker).StackTrace()
	if len(stack1.Frames()) == len(stack2.Frames()) {
		t.Error("stacks same")
	}

	var stack pruntime.Stack

	// nil error should return nil stack
	stack = errorglue.GetInnermostStack(nil)
	if stack != nil {
		t.Errorf("stack not nil")
	}

	// error without stack should return nil
	stack = errorglue.GetInnermostStack(errNoStack)
	if stack != nil {
		t.Errorf("stack not nil")
	}

	stack = errorglue.GetInnermostStack(errOneStack)
	if len(stack.Frames()) == 0 {
		t.Error("stack3 len 0")
	}
	if len(stack.Frames()) != len(stack1.Frames()) {
		t.Errorf("stack3 bad length %d exp %d", len(stack.Frames()), len(stack1.Frames()))
	}

	stack = errorglue.GetInnermostStack(errTwoStacks)
	if len(stack.Frames()) == 0 {
		t.Error("stack4 len 0")
	}
	if len(stack.Frames()) != len(stack1.Frames()) {
		t.Errorf("stack4 bad length %d exp %d", len(stack.Frames()), len(stack1.Frames()))
	}
}
