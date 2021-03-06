/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package goid

import (
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pruntime"
)

func TestNewStack(t *testing.T) {
	expectedFrameLength := 2
	stack := NewStack(0)
	if stack == nil {
		t.Errorf("NewStack nil return")
	}
	actualLength := len(stack.Frames())
	if actualLength != expectedFrameLength {
		t.Errorf("Bad stack.Frames length %d expected %d", actualLength, expectedFrameLength)
	}
	if stack.Creator().Line == 0 {
		t.Errorf("NewStack failed")
	}
	// goid.Frame{
	//  	CodeLocation:pruntime.CodeLocation{
	// 		File:"/opt/sw/parl/goid/stack_test.go",
	//		Line:12,
	//		FuncName:"github.com/haraldrudell/parl/goid.TestNewStack"
	//	},
	//	Args:"(0x1400011a340)"
	// }
	cl := pruntime.NewCodeLocation(0)
	if stack.Frames()[0].Loc().File != cl.File {
		t.Errorf("Bad file: %s expected %s", stack.Frames()[0].Loc().File, cl.File)
	}
}

func TestParseFirstStackLine(t *testing.T) {
	input := "goroutine 19 [running]:\ngarbage"
	expID := parl.ThreadID("19")
	expStatus := parl.ThreadStatus("running")

	ID, status, err := ParseFirstLine(input)
	if err != nil {
		t.Errorf("ParseFirstStackLine err: %v", err)
	}
	if ID != expID {
		t.Errorf("ID: %q exp: %q", ID, expID)
	}
	if status != expStatus {
		t.Errorf("status: %q exp: %q", status, expStatus)
	}
}
