/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"testing"

	"github.com/haraldrudell/parl/pdebug"
)

func TestThreadDataWrap(t *testing.T) {
	stack := pdebug.NewStack(0)

	var threadDataWrap ThreadDataWrap
	var threadData *ThreadData
	var isValid bool

	threadDataWrap = ThreadDataWrap{}
	if threadDataWrap.HaveThreadID() {
		t.Error("haveThreadID true")
	}

	threadDataWrap.Update(stack)
	if threadDataWrap.ThreadID() != stack.ID() {
		t.Errorf("bad ID %q exp %q", threadDataWrap.ThreadID(), stack.ID())
	}
	threadData, isValid = threadDataWrap.Get()
	_ = threadData
	_ = isValid
	threadDataWrap.SetCreator(stack.Creator())
}
