/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/goid"
	"github.com/haraldrudell/parl/pruntime"
)

func TestGo(t *testing.T) {
	var err = errors.New("x")
	var label = "label"

	var goGroup parl.GoGroup
	var g0 parl.Go
	var g0Impl *Go
	var subGroup parl.SubGroup
	var subGo parl.SubGo
	var goError parl.GoError
	var ok bool
	var ctx0, ctx context.Context

	// Register() AddError() Go() SubGo() SubGroup() Done() ThreadInfo()
	// GoID() Wait()
	goGroup = NewGoGroup(context.Background())
	g0 = goGroup.Go()
	g0Impl = g0.(*Go)
	if g0Impl.isTerminated.IsTrue() {
		t.Error("Go terminated")
	}
	g0.Register(label)
	if !g0.GoID().IsValid() {
		t.Error("Go.GoID bad")
	}
	if g0.ThreadInfo().Name() != label {
		t.Error("g0.Register bad")
	}
	if subGroup = g0.SubGroup(); subGroup == nil {
		t.Error("Go.SubGroup bad")
	}
	if subGo = g0.SubGo(); subGo == nil {
		t.Error("Go.SubGo bad")
	}
	g0.AddError(err)
	if goError = <-goGroup.Ch(); !errors.Is(goError.Err(), err) ||
		goError.ErrContext() != parl.GeNonFatal {
		t.Error("g0.AddError bad")
	}
	g0.Done(&err)
	if goError = <-goGroup.Ch(); !errors.Is(goError.Err(), err) ||
		goError.ErrContext() != parl.GeExit {
		t.Error("g0.Done bad")
	}
	if _, ok = <-goGroup.Ch(); ok {
		t.Error("g0.Done goGroup errch did not close")
	}
	if !g0Impl.wg.IsZero() {
		t.Error("!g0Impl.wg.IsZero")
		t.FailNow()
	}
	g0.Wait()

	// Cancel() Context()
	goGroup = NewGoGroup(context.Background())
	ctx0 = goGroup.Context()
	g0 = goGroup.Go()
	g0.Cancel()
	if ctx = g0.Context(); ctx.Err() == nil {
		t.Error("Go.Cancel did not cancel context")
	}
	if ctx0 != ctx {
		t.Error("Go.Context bad")
	}

	// String(): "go:34:testing.(*T).Run()-testing.go:1629"
	t.Log(g0.String())
	//t.Fail()
}

type T struct {
	g0                parl.Go
	label             string
	cL                *pruntime.CodeLocation
	goroutineThreadID *parl.ThreadID
	wg                sync.WaitGroup
}

// tt.g must have Register and NewCodeLocation on same line
//   - tt.g is the goFunction of a goroutine
func (tt *T) g() { tt.g0.Register(tt.label); tt.h(pruntime.NewCodeLocation(0)) }
func (tt *T) h(g0RegisterLine *pruntime.CodeLocation) {
	defer tt.wg.Done()

	*tt.cL = *g0RegisterLine
	*tt.goroutineThreadID = goid.GoID()
}

func TestGo_Frames(t *testing.T) {
	var expLabel = "LABEL"
	var expThreadID = goid.GoID()
	var cL *pruntime.CodeLocation
	var g0 parl.Go
	var subGo parl.SubGo
	var subGroup parl.SubGroup
	var goFunctionCL pruntime.CodeLocation
	var goroutineThreadID parl.ThreadID

	var goGroup parl.GoGroup = NewGoGroup(context.Background())
	var parentGo parl.Go = goGroup.Go()

	// Go.Go(): Go.String() has caller location
	//	- the location is stored in the thread field
	// g0 and cL assignments on same line
	g0, cL = parentGo.Go(), pruntime.NewCodeLocation(0)
	if !strings.HasSuffix(g0.String(), cL.Short()) {
		var _ = (&Go{}).String
		var _ ThreadData
		t.Errorf("Go.Go: Go.String BAD: %q exp suffix: %q", g0.String(), cL.Short())
	}

	// Go.Go(): before g0.Register(), there is an intermediate createLocation
	//	- this has only creator set
	id, cre := g0.Creator()
	var _ ThreadData
	// Go.Go() before g0.Register: Go.ThreadInfo: cre:g0.TestGo_Frames()-go_test.go:113 creatorID: 4
	t.Logf("Go.Go() before g0.Register: Go.ThreadInfo: %s creatorID: %s", g0.ThreadInfo(), id)
	if id != expThreadID {
		t.Errorf("Go.CreatorID: %q exp %q", id, expThreadID)
	}
	if cre.Short() != cL.Short() {
		t.Errorf("Go.Creator Short: %q exp %q", cre.Short(), cL.Short())
	}

	// Go.Register(): updates ThreadID goFuncLocation and label
	tt := T{g0: g0, label: expLabel, cL: &goFunctionCL, goroutineThreadID: &goroutineThreadID}
	tt.wg.Add(1)
	go tt.g()
	tt.wg.Wait()
	var threadData = g0.ThreadInfo()
	if threadData.ThreadID() != goroutineThreadID {
		t.Errorf("Go ThreadID: %q exp %q", threadData.ThreadID(), goroutineThreadID)
	}
	t.Logf("goFunction: %s", threadData.Func().Dump())
	t.Logf("goFunctionCL: %s", goFunctionCL.Dump())
	if threadData.Func().Short() != goFunctionCL.Short() {
		t.Errorf("Go goFunction Short: %q exp %q", threadData.Func().Short(), goFunctionCL.Short())
	}
	if threadData.Name() != expLabel {
		t.Errorf("Go thread label: %q exp %q", threadData.Name(), expLabel)
	}

	// g0.SubGo() g0.SubGroup() have invoking creator location
	var _ parl.Go
	subGo, cL = g0.SubGo(), pruntime.NewCodeLocation(0)
	if !strings.HasSuffix(subGo.String(), cL.Short()) {
		t.Errorf("g0.SubGo() invoker location: %q expPrefix %q", subGo.String(), cL.Short())
	}
	subGroup, cL = g0.SubGroup(), pruntime.NewCodeLocation(0)
	if !strings.HasSuffix(subGroup.String(), cL.Short()) {
		t.Errorf("g0.SubGroup() invoker location: %q expPrefix %q", subGroup.String(), cL.Short())
	}
}
