/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package goid

import (
	"runtime/debug"
	"strings"
	"testing"
)

func TestGoRoutineID(t *testing.T) {
	stackIndex := 10
	var expectedID ThreadID
	s := string(debug.Stack())[stackIndex:]
	if index := strings.Index(s, "\x20"); index == -1 {
		t.Error("debug.Stack failed")
	} else {
		expectedID = ThreadID(s[:index])
	}
	actual := GoID()
	if actual != expectedID {
		t.Errorf("GoRoutineID bad: %q expected %q", actual, expectedID)
	}
}
