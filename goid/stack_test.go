/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package goid

import "testing"

func TestNewStack(t *testing.T) {
	expectedFrameLength := 2
	stack := NewStack(0)
	if stack == nil {
		t.Errorf("NewStack nil return")
	}
	actualLength := len(stack.Frames)
	if actualLength != expectedFrameLength {
		t.Errorf("Bad stack.Frames length %d expected %d", actualLength, expectedFrameLength)
	}
	if stack.Creator.Line == 0 {
		t.Errorf("NewStack failked")
	}
	t.Errorf("\n%#v", stack.Frames[0])
}
