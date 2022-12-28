/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"testing"

	"github.com/haraldrudell/parl/pdebug"
)

func TestThreadData(t *testing.T) {
	stack := pdebug.NewStack(0)

	var threadData ThreadData
	_ = threadData

	threadData = ThreadData{}

	threadData.Update(stack)
	if threadData.ThreadID() != stack.ID() {
		t.Errorf("bad ID %q exp %q", threadData.ThreadID(), stack.ID())
	}
	threadID, createLocation, funcLocation := threadData.Get()
	if threadID != stack.ID() {
		t.Errorf("bad ID2 %q exp %q", threadID, stack.ID())
	}
	_ = createLocation
	_ = funcLocation
	threadData.SetCreator(stack.Creator())
}
