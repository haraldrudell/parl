/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pruntime"
)

func TestGoGroup(t *testing.T) {
	messageBad := "bad"
	errBad := errors.New(messageBad)
	var label = "label"

	var goGroup parl.GoGroup
	var goGroupImpl *GoGroup
	var g0 parl.Go
	var goImpl *Go
	var goError parl.GoError
	var goError2 parl.GoError
	var ok bool
	var count int
	var subGo parl.SubGo
	var subGroup parl.SubGroup
	var ctx0, ctx context.Context
	var threads []parl.ThreadData
	var fatals int
	var onFirstFatal = func(goGen parl.GoGen) { fatals++ }
	var expectG0ID uint64

	// g0.NewGoGroup returns *g0.GoGroup: NewGoGroup()
	goGroup = NewGoGroup(context.Background())
	if goGroupImpl, ok = goGroup.(*GoGroup); !ok {
		t.Error("NewGoGroup did not return *g0.GoGroup")
		t.FailNow()
	}
	goGroup.SetDebug(parl.AggregateThread)

	// fail thread exit: Go() GoDone() Ch() Threads() NamedThreads() IsEnd() Wait()
	g0 = goGroup.Go()
	if goImpl, ok = g0.(*Go); !ok {
		t.Error("GoGroup.Go() did not return *g0.Go")
		t.FailNow()
	}
	_ = goImpl
	g0.Register("x")
	if len(goGroup.Threads()) != 1 {
		t.Error("goGroup.Threads length bad")
	}
	if len(goGroup.NamedThreads()) != 1 {
		t.Error("goGroup.NamedThreads length bad")
	}
	goGroupImpl.GoDone(g0, errBad)
	// verify that error channel has error
	if count = goGroupImpl.ch.Count(); count != 1 {
		t.Errorf("bad Ch Count: %d exp 1", count)
	}
	goError, ok = <-goGroup.Ch() // receive errBad
	if !ok {
		t.Error("goGroup.Ch closed")
	}
	if goError == nil {
		t.Error("goError nil")
		t.FailNow()
	}
	if !errors.Is(goError.Err(), errBad) {
		t.Errorf("wrong error: %q %x exp %q %x", goError.Error(), goError, errBad.Error(), errBad)
	}
	// verify GoGroup termination
	_, ok = <-goGroup.Ch() // check that channel is now closed
	if ok {
		t.Error("goGroup.Ch did not close")
	}
	// the channel managed by NBChan goGroup.ch is closed
	//	- the channel is goGroup.NBChan.ClosableChan.ch
	//	- that means NBChan.sendThreadDefer did NBChan.Close
	//
	// - goGroup.isEnd does NBChan.IsClosed delegating to:
	//	- ClosableChan.IsClosed which returns
	//	- parl.Once.IsDone.IsTrue()
	//	- there is a race condition between closing the channel
	//	- and setting isDone true
	//	- during which isEnd will indicate not ended
	//
	// therefore wait for the known pending close here:
	goGroupImpl.ch.WaitForClose()

	if !goGroupImpl.isEnd() {
		t.Error("goGroup did not terminate")
	}
	if !goGroupImpl.wg.IsZero() {
		t.Error("goGroup wg not zero")
		t.FailNow()
	}
	goGroup.Wait()

	// ConsumeError() EnableTermination() CascadeEnableTermination()
	// IsEnableTermination()
	goGroup = NewGoGroup(context.Background())
	goGroupImpl = goGroup.(*GoGroup)
	g0 = goGroup.Go()
	goError2 = NewGoError(errBad, parl.GeNonFatal, g0)
	goGroupImpl.ConsumeError(goError2)
	var goErrorActual = <-goGroup.Ch() // GoError from ConsumeError
	if goErrorActual != goError2 {
		t.Error("GoGroup.ConsumeError failed")
	}
	if goGroupImpl.isEnd() {
		t.Error("1 GoGroup terminated")
	}
	if !goGroup.IsEnableTermination() {
		t.Error("IsEnableTermination false")
	}
	goGroup.EnableTermination(false)
	if goGroup.IsEnableTermination() {
		t.Error("IsEnableTermination true")
	}
	goGroupImpl.GoDone(g0, nil)
	<-goGroup.Ch() // GoError from GoDone
	if goGroupImpl.isEnd() {
		t.Error("2 GoGroup terminated")
	}
	goGroup.EnableTermination(true)
	if !goGroupImpl.isEnd() {
		t.Error("GoGroup did not terminate")
	}

	// SubGo() SubGroup()
	goGroup = NewGoGroup(context.Background())
	subGo = goGroup.SubGo()
	goGroupImpl = subGo.(*GoGroup)
	if goGroupImpl.isSubGroup.Load() {
		t.Error("SubGo returned SubGroup")
	}
	if goGroupImpl.hasErrorChannel.Load() {
		t.Error("SubGo has error channel")
	}
	subGroup = goGroup.SubGroup()
	goGroupImpl = subGroup.(*GoGroup)
	if !goGroupImpl.isSubGroup.Load() {
		t.Error("SubGroup did not return SubGroup")
	}
	if !goGroupImpl.hasErrorChannel.Load() {
		t.Error("SubGroup does not have error channel")
	}

	// Context()
	ctx0 = parl.NewCancelContext(context.Background())
	goGroup = NewGoGroup(ctx0)
	parl.InvokeCancel(ctx0)
	ctx = goGroup.Context()
	if ctx.Err() == nil {
		t.Error("goGroup context did not cancel from parent context")
	}

	// Cancel()
	goGroup = NewGoGroup(context.Background())
	goGroup.Cancel()
	if ctx = goGroup.Context(); ctx.Err() == nil {
		t.Error("goGroup cancel did not cancel context")
	}

	// Add() UpdateThread() SetDebug() Threads() NamedThreads() G0ID()
	goGroup = NewGoGroup(context.Background())
	goGroupImpl = goGroup.(*GoGroup)
	expectG0ID = uint64(goGroupImpl.goEntityID.id)
	if expectG0ID != uint64(goGroupImpl.G0ID()) {
		t.Error("goGroupImpl.G0ID bad")
	}
	goGroup.Go().Register()
	if goGroupImpl.wg.Count() != 1 {
		t.Errorf("goGroupImpl.wg.Count not 1: %d", goGroupImpl.wg.Count())
	}
	if len(goGroup.Threads()) > 0 {
		t.Error("goGroup no-debug collects threads")
	}
	goGroup.SetDebug(parl.DebugPrint)
	if !goGroupImpl.isDebug.Load() {
		t.Error("goGroup.SetDebug DebugPrint failed")
	}
	goGroup.SetDebug(parl.AggregateThread)
	if goGroupImpl.isDebug.Load() {
		t.Error("goGroup.SetDebug AggregateThread failed")
	}
	goGroup.Go().Register(label)
	if len(goGroup.Threads()) != 1 {
		t.Errorf("goGroup.Threads not 1: %d", len(goGroup.Threads()))
	}
	threads = goGroup.NamedThreads()
	if len(threads) != 1 || threads[0].Name() != label {
		t.Error("goGroup.NamedThreads bad")
	}

	// FirstFatal()
	goGroup = NewGoGroup(context.Background(), onFirstFatal)
	goGroup.Go().Done(&errBad)
	if fatals == 0 {
		t.Error("onFirstFatal bad")
	}
}

func TestGoGroup_Frames(t *testing.T) {
	// goGroup and cL on same line
	var goGroup parl.GoGroup
	var subGo parl.SubGo
	var subGroup parl.SubGroup
	var g0 parl.Go
	var cL *pruntime.CodeLocation

	// NewGoGroup: GoGroup.String() includes NewGoGroup caller location
	goGroup, cL = NewGoGroup(context.Background()), pruntime.NewCodeLocation(0)
	if !strings.HasSuffix(goGroup.String(), cL.Short()) {
		t.Errorf("GoGroup.String BAD: %q exp suffix: %q", goGroup.String(), cL.Short())
	}

	// GoGroup.SubGo includes caller location
	subGo, cL = goGroup.SubGo(), pruntime.NewCodeLocation(0)
	if !strings.HasSuffix(subGo.String(), cL.Short()) {
		t.Errorf("SubGo.String: %q exp suffix: %q", subGo.String(), cL.Short())
	}

	// GoGroup.SubGroup includes caller location
	subGroup, cL = goGroup.SubGroup(), pruntime.NewCodeLocation(0)
	if !strings.HasSuffix(subGroup.String(), cL.Short()) {
		t.Errorf("SubGroup.String: %q exp suffix: %q", subGroup.String(), cL.Short())
	}

	// GoGroup.Go includes caller location
	g0, cL = goGroup.Go(), pruntime.NewCodeLocation(0)
	if !strings.HasSuffix(g0.String(), cL.Short()) {
		var _ = (&GoGroup{}).Go
		// Go.String: "subGroup#3_threads:0(0)_New:g0.TestGoGroup_Frames()-go-group_test.go:217" exp suffix: "g0.TestGoGroup_Frames()-go-group_test.go:223"
		t.Errorf("Go.String: %q exp suffix: %q", subGroup.String(), cL.Short())
	}
}

func TestSubGo(t *testing.T) {
	var err = errors.New("bad")

	var goGroup parl.GoGroup
	var goGroupImpl, subGoImpl *GoGroup
	var subGo parl.SubGo
	var goError, goError2 parl.GoError
	var parlGo parl.Go
	var ok bool

	// SubGo non-fatal error
	goGroup = NewGoGroup(context.Background())
	goGroupImpl = goGroup.(*GoGroup)
	subGo = goGroup.SubGo()
	subGoImpl = subGo.(*GoGroup)
	goError = NewGoError(err, parl.GeNonFatal, nil)
	subGoImpl.ConsumeError(goError)
	// the non-fatal subGo error should be recevied on GoGroup error channel
	goError2 = <-goGroup.Ch()
	if goError2 != goError {
		t.Errorf("bad non-fatal subgo error")
	}

	// SubGo fatal thread termination
	parlGo = subGo.Go()
	parlGo.Done(&err)
	// the SubGo fatal error should be recevied on GoGroup error channel
	goError2 = <-goGroup.Ch()
	if !errors.Is(goError2.Err(), err) {
		t.Error("bad fatal subgo error")
	}
	// subgo should now terminate after its only thread exited
	if !subGoImpl.isEnd() {
		t.Error("subGo did not terminate")
	}

	// gogroup should now have terminated and closed its error channel
	//	- its only thread did exit
	goError2, ok = <-goGroup.Ch() // wait for subGroup channel to close
	if ok {
		t.Errorf("goGroup channel did not close: %s", goError2)
	}
	if !goGroupImpl.isEnd() {
		t.Error("goGroup did not terminate")
	}
}

func TestSubGroup(t *testing.T) {
	var err = errors.New("bad")

	var goGroup parl.GoGroup
	var goGroupImpl, subGroupImpl *GoGroup
	var subGroup parl.SubGroup
	var goError, goError2 parl.GoError
	var parlGo parl.Go
	var ok bool

	// non-fatal error: sent to gogroup
	goGroup = NewGoGroup(context.Background())
	goGroupImpl = goGroup.(*GoGroup)
	subGroup = goGroup.SubGroup()
	subGroupImpl = subGroup.(*GoGroup)
	goError = NewGoError(err, parl.GeNonFatal, nil)
	subGroupImpl.ConsumeError(goError)
	goError2 = <-goGroup.Ch()
	if goError2 != goError {
		t.Errorf("bad non-fatal subgroup error")
	}

	// fatal error:
	//	- a thread exits with g0.Done having error
	//	- the subGroup hides the fatal error from the parent
	//	- the parent receives non-fatal GeLocalChan of the error and a GeExit with no error
	//	- subgroup emits fatal error on its error channel
	parlGo = subGroupImpl.Go()
	parlGo.Done(&err)
	// goGroup GeLocalChan
	goError2 = <-goGroup.Ch()
	if !errors.Is(goError2.Err(), err) {
		t.Error("bad gogroup error")
	}
	if goError2.ErrContext() != parl.GeLocalChan {
		t.Errorf("bad gogroup error context: %s", goError2.ErrContext())
	}
	// goGroup good thread exit
	goError2 = <-goGroup.Ch()
	if goError2.Err() != nil {
		t.Errorf("bad gogroup error: %s", goError2.String())
	}
	if goError2.ErrContext() != parl.GeExit {
		t.Errorf("bad gogroup error context: %s", goError2.ErrContext())
	}
	// SubGroup: GeExit fatal error
	goError2 = <-subGroup.Ch()
	if !errors.Is(goError2.Err(), err) {
		t.Error("bad fatal subgroup error")
	}

	// subgroup should now exit:
	goError2, ok = <-subGroup.Ch() // wait for subGroup channel to close
	if ok {
		t.Errorf("subGroup channel did not close: %s", goError2)
	}
	if !subGroupImpl.isEnd() {
		t.Error("subGroup did not terminate")
	}

	// gogroup exits
	goError2, ok = <-goGroup.Ch() // wait for subGroup channel to close
	if ok {
		t.Errorf("goGroup channel did not close: %s", goError2)
	}
	if !goGroupImpl.isEnd() {
		t.Error("goGroup did not terminate")
	}
}

func TestCancel(t *testing.T) {
	var ctx = parl.AddNotifier(context.Background(), func(slice pruntime.StackSlice) {
		t.Logf("ALLCANCEL %s", slice)
	})

	var threadGroup = NewGoGroup(ctx)
	threadGroup.(*GoGroup).AddNotifier(func(slice pruntime.StackSlice) {
		t.Logf("CANCEL %s %s", GoChain(threadGroup), slice)
	})
	var subGroup = threadGroup.SubGroup()
	subGroup.(*GoGroup).AddNotifier(func(slice pruntime.StackSlice) {
		t.Logf("CANCEL %s %s", GoChain(subGroup), slice)
	})
	t.Logf("STATE0: %t %t", threadGroup.Context().Err() != nil, subGroup.Context().Err() != nil)
	if threadGroup.Context().Err() != nil {
		t.Error("threadGroup canceled")
	}
	if subGroup.Context().Err() != nil {
		t.Error("subGroup canceled")
	}
	subGroup.Cancel()
	t.Logf("STATE1: %t %t", threadGroup.Context().Err() != nil, subGroup.Context().Err() != nil)
	if threadGroup.Context().Err() != nil {
		t.Error("threadGroup canceled")
	}
	if subGroup.Context().Err() == nil {
		t.Error("subGroup did not cancel")
	}
	//t.Fail()
}
