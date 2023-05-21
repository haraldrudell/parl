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

	var threadDataWrap ThreadSafeThreadData
	var threadData *ThreadData

	threadDataWrap = ThreadSafeThreadData{}
	if threadDataWrap.HaveThreadID() {
		t.Error("haveThreadID true")
	}

	threadDataWrap.Update(
		stack.ID(),
		stack.Creator(),
		stack.GoFunction(),
		"",
	)
	if threadDataWrap.ThreadID() != stack.ID() {
		t.Errorf("bad ID %q exp %q", threadDataWrap.ThreadID(), stack.ID())
	}
	threadData = threadDataWrap.Get()
	_ = threadData
	threadDataWrap.SetCreator(stack.Creator())
}
