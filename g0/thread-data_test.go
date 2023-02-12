/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"strings"
	"sync"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pdebug"
)

const (
	threadDataLabel = "myThreadName"
)

func TestThreadData(t *testing.T) {
	var createShort = ".(*SomeType).SomeCode()"
	var funcShort = ".(*SomeType).SomeFunction()"

	var exp string

	// check zero-value results
	var threadData ThreadData
	if threadData.ThreadID().IsValid() {
		t.Error("threadData.ThreadID().IsValid() true")
	}
	if threadData.Create().IsSet() {
		t.Error("threadData.Create().IsSet() true")
	}
	if threadData.Func().IsSet() {
		t.Error("threadData.Func().IsSet() true")
	}
	if name := threadData.Name(); name != "" {
		t.Errorf("threadData.Name() not empty: %q", name)
	}
	if short := threadData.Short(); short != ThreadDataEmpty {
		t.Errorf("threadData.Short() zero-value: %q exp %q", short, ThreadDataEmpty)
	}
	if short := threadData.String(); short != ThreadDataEmpty {
		t.Errorf("threadData.String() zero-value: %q exp %q", short, ThreadDataEmpty)
	}

	// check nil returns
	var threadDatap *ThreadData
	if short := threadDatap.Short(); short != ThreadDataNil {
		t.Errorf("nil.Short(): %q exp %q", short, ThreadDataNil)
	}

	// check populated returns
	var someType SomeType
	someType.SomeCode(&threadData)
	threadID, createLocation, funcLocation, label := threadData.Get()
	if threadData.ThreadID() != threadID {
		t.Error("bad threadData.threadID()")
	}
	if *threadData.Create() != createLocation {
		t.Error("bad threadData.Create()")
	}
	if *threadData.Func() != funcLocation {
		t.Error("bad threadData.Func()")
	}
	if threadData.Name() != label {
		t.Error("bad threadData.Name()")
	}
	if threadID != someType.stack.ID() {
		t.Errorf("bad ID %q exp %q", threadID, someType.stack.ID())
	}
	if !strings.Contains(createLocation.Short(), createShort) {
		t.Errorf("createLocation.Short(): %q exp %q", createLocation.Short(), createShort)
	}
	if !strings.Contains(funcLocation.Short(), funcShort) {
		t.Errorf("funcLocation.Short(): %q exp %q", funcLocation.Short(), funcShort)
	}
	if label != threadDataLabel {
		t.Errorf("label: %q exp %q", label, threadDataLabel)
	}
	exp = label + ":" + threadID.String()
	if short := threadData.Short(); short != exp {
		t.Errorf("threadData.Short(): %q exp %q", short, exp)
	}
	if short := threadData.String(); !strings.Contains(short, exp) {
		t.Errorf("threadData.String(): %q exp %q", short, exp)
	}

	threadData.SetCreator(someType.stack.Creator())
}

// ITEST= go test -v -run '^TestThreadDataValues$' ./g0
func TestThreadDataValues(t *testing.T) {

	var threadData ThreadData
	t.Logf("zero-value ThreadID: %q", threadData.ThreadID())
	t.Logf("zero-value Create().Short(): %q", threadData.Create().Short())
	t.Logf("zero-value Name(): %q", threadData.Name())
	t.Logf("zero-value Short(): %q", threadData.Name())
	t.Logf("zero-value String(): %q", &threadData)

	var someType SomeType
	someType.SomeCode(&threadData)
	t.Logf("ThreadID: %q", threadData.ThreadID())
	t.Logf("Create().Short(): %q", threadData.Create().Short())
	t.Logf("Func().Short(): %q", threadData.Func().Short())
	t.Logf("Name(): %q", threadData.Name())
	t.Logf("Short(): %q", threadData.Short())
	t.Logf("String(): %q", &threadData)

}

type SomeType struct {
	wg    sync.WaitGroup
	stack parl.Stack
}

func (st *SomeType) SomeCode(threadData *ThreadData) {
	st.wg.Add(1)
	go st.SomeFunction(threadData)
	st.wg.Wait()
}
func (st *SomeType) SomeFunction(threadData *ThreadData) {
	defer st.wg.Done()

	st.SomeMethod(threadData)
}
func (st *SomeType) SomeMethod(threadData *ThreadData) {
	st.stack = pdebug.NewStack(0)
	threadData.Update(st.stack, threadDataLabel)
}
