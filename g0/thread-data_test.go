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

func TestThreadDataZeroValue(t *testing.T) {

	// check zero-value results
	var threadData ThreadData
	// a zero-value threadData should not be valid
	if threadData.ThreadID().IsValid() {
		t.Error("threadData.ThreadID().IsValid() true")
	}
	// a zero-value threadData should not have Create
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
}

func TestThreadData(t *testing.T) {
	var expCreateShort = ".(*SomeType).SomeCode"
	var expFuncShort = ".(*SomeType).SomeFunction"

	var exp string
	var threadData ThreadData

	// check populated returns
	var someType SomeType
	someType.SomeCode(&threadData)

	// ID: 36 IsMain: false status: running
	// github.com/haraldrudell/parl/g0.(*SomeType).SomeMethod(0x140001482c0, 0x0?)
	// 	thread-data_test.go:135
	// github.com/haraldrudell/parl/g0.(*SomeType).SomeFunction(0x14000106b60?, 0x102ef8280?)
	// 	thread-data_test.go:132
	// cre: github.com/haraldrudell/parl/g0.(*SomeType).SomeCode in goroutine 35-thread-data_test.go:126
	//t.Logf("stack: \n%s", someType.stack)

	// ThreadID: 20
	// Create: File: "/opt/sw/parl/g0/thread-data_test.go" Line: 147 FuncName: "github.com/haraldrudell/parl/g0.(*SomeType).SomeCode in goroutine 19"
	// Func: File: "/opt/sw/parl/g0/thread-data_test.go" Line: 153 FuncName: "github.com/haraldrudell/parl/g0.(*SomeType).SomeFunction"
	// Name: myThreadName
	// t.Logf("\n"+
	// 	"ThreadID: %s\n"+
	// 	"Create: %s\n"+
	// 	"Func: %s\n"+
	// 	"Name: %s",
	// 	threadData.ThreadID().String(),
	// 	threadData.Create().Dump(),
	// 	threadData.Func().Dump(),
	// 	threadData.Name(),
	// )

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
	if !strings.Contains(createLocation.Short(), expCreateShort) {
		t.Errorf("createLocation.Short(): %q exp %q", createLocation.Short(), expCreateShort)
	}
	if !strings.Contains(funcLocation.Short(), expFuncShort) {
		t.Errorf("funcLocation.Short(): %q exp %q", funcLocation.Short(), expFuncShort)
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
	var creator, _, _ = someType.stack.Creator()
	threadData.SetCreator(creator)
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

func (s *SomeType) SomeCode(threadData *ThreadData) {
	s.wg.Add(1)
	go s.SomeFunction(threadData)
	s.wg.Wait()
}
func (s *SomeType) SomeFunction(threadData *ThreadData) {
	defer s.wg.Done()

	s.SomeMethod(threadData)
}
func (s *SomeType) SomeMethod(threadData *ThreadData) {
	s.stack = pdebug.NewStack(0)
	var creator, _, _ = s.stack.Creator()
	threadData.Update(
		s.stack.ID(),
		creator,
		s.stack.GoFunction(),
		threadDataLabel)
}
