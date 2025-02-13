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
	const (
		label = "label"
	)
	var (
		err = errors.New("x")
	)

	var (
		goGroup       parl.GoGroup
		goGroupImpl   *GoGroup
		goErrorSource parl.Source1[parl.GoError]
		goImpl        *Go
		subGroup      parl.SubGroup
		subGo         parl.SubGo
		goError       parl.GoError
		ok            bool
		ctx0, ctx     context.Context
	)

	// Register() AddError() Go() SubGo() SubGroup() Done() Wait()
	// WaitCh() Cancel() Context() ThreadInfo() Creator()
	// GoRoutine() GoID() EntityID()
	var g parl.Go
	var reset = func() {
		goGroup = NewGoGroup(context.Background())
		goGroupImpl = goGroup.(*GoGroup)
		goErrorSource = &goGroupImpl.goErrorStream
		g = goGroup.Go()
		goImpl = g.(*Go)
	}

	// Register() AddError() Go() SubGo() SubGroup() Done() ThreadInfo()
	// GoID() Wait()
	reset()
	if goImpl.endCh.IsClosed() {
		t.Error("Go created terminated")
	}
	g.Register(label)
	if !g.GoID().IsValid() {
		t.Error("Go.GoID bad")
	}
	if g.ThreadInfo().Name() != label {
		t.Error("Go.Register bad")
	}
	if subGroup = g.SubGroup(); subGroup == nil {
		t.Error("Go.SubGroup bad")
	}
	if subGo = g.SubGo(); subGo == nil {
		t.Error("Go.SubGo bad")
	}
	g.AddError(err)
	if goError, _ = parl.AwaitValue(goErrorSource); !errors.Is(goError.Err(), err) ||
		goError.ErrContext() != parl.GeNonFatal {
		t.Error("g0.AddError bad")
	}
	g.Done(&err)
	if goError, _ = parl.AwaitValue(goErrorSource); !errors.Is(goError.Err(), err) ||
		goError.ErrContext() != parl.GeExit {
		t.Error("g0.Done bad")
	}
	if _, ok = parl.AwaitValue(goErrorSource); ok {
		t.Error("g0.Done goGroup errch did not close")
	}
	g.Wait()

	// Cancel() Context()
	reset()
	ctx0 = goGroup.Context()
	g.Cancel()
	if ctx = g.Context(); ctx.Err() == nil {
		t.Error("Go.Cancel did not cancel context")
	}
	if ctx0 != ctx {
		t.Error("Go.Context bad")
	}

	// String(): "go:34:testing.(*T).Run()-testing.go:1629"
	t.Log(g.String())
}

func TestGo_Frames(t *testing.T) {
	const (
		expLabel = "LABEL"
	)
	var (
		expThreadID = goid.GoID()
	)

	var (
		// goGroup provides Go objects
		goGroup = NewGoGroup(context.Background())
		// parentGo is a fake Go that generates the Go used by the goroutine
		parentGo          parl.Go = goGroup.Go()
		goInvocationLine  *pruntime.CodeLocation
		g                 parl.Go
		subGo             parl.SubGo
		subGroup          parl.SubGroup
		goFunctionCL      *pruntime.CodeLocation
		goroutineThreadID parl.ThreadID
		regTester         *registerTester
		creatorThreadID   parl.ThreadID
		createLocation    *pruntime.CodeLocation
	)

	// Go.Go(): Go.String() should have caller location
	//	- the location is stored in the thread field
	//	- generate g and the code-line where it was created
	//	- g and goInvocationLine assignments on same line
	g, goInvocationLine = parentGo.Go(), pruntime.NewCodeLocation(0)
	if s := g.String(); !strings.HasSuffix(s, goInvocationLine.Short()) {
		var _ = (&Go{}).String
		var _ ThreadData
		t.Errorf("Go.Go: Go.String BAD: %q exp suffix: %q", s, goInvocationLine.Short())
	}

	// Go.Creator() should be initialized with creating thread first
	//	- Go.Go(): before g0.Register(), there is an intermediate createLocation
	//	- this has only creator set
	creatorThreadID, createLocation = g.Creator()
	var _ ThreadData
	// Go.Go() before g0.Register: Go.ThreadInfo: cre:g0.TestGo_Frames()-go_test.go:113 creatorID: 4
	t.Logf("Go.Go() before g0.Register: Go.ThreadInfo: %s creatorID: %s", g.ThreadInfo(), creatorThreadID)
	if creatorThreadID != expThreadID {
		t.Errorf("Go.CreatorID: %q exp %q", creatorThreadID, expThreadID)
	}
	if createLocation.Short() != goInvocationLine.Short() {
		t.Errorf("Go.Creator Short: %q exp %q", createLocation.Short(), goInvocationLine.Short())
	}

	// Go.Register() should update ThreadID goFuncLocation and label
	regTester = newRegisterTester(g, expLabel)
	go regTester.goMethod().goFunction()
	goFunctionCL, goroutineThreadID = regTester.result()
	var threadData = g.ThreadInfo()
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

	// g0.SubGo() should have invoking creator location
	var _ parl.Go
	subGo, goInvocationLine = g.SubGo(), pruntime.NewCodeLocation(0)
	if !strings.HasSuffix(subGo.String(), goInvocationLine.Short()) {
		t.Errorf("g0.SubGo() invoker location: %q expPrefix %q", subGo.String(), goInvocationLine.Short())
	}

	// g0.SubGroup() should have invoking creator location
	subGroup, goInvocationLine = g.SubGroup(), pruntime.NewCodeLocation(0)
	if !strings.HasSuffix(subGroup.String(), goInvocationLine.Short()) {
		t.Errorf("g0.SubGroup() invoker location: %q expPrefix %q", subGroup.String(), goInvocationLine.Short())
	}
}

// registerTester tests Go.Register method
type registerTester struct {
	// g is the Go object managing the goFunction goroutine
	g parl.Go
	// label is the thread-name the goroutine registers as
	label string
	// the code line from the goroutine
	cL *pruntime.CodeLocation
	// goroutineThreadID is the thread ID for the launched goroutine
	goroutineThreadID parl.ThreadID
	// wg makes the goroutine awaitable
	wg sync.WaitGroup
}

// newRegisterTester returns a tester for [Go.Regsiter]
func newRegisterTester(g parl.Go, label string) (r *registerTester) {
	return &registerTester{
		g:     g,
		label: label,
	}
}

// goMethod returns the method to use in go statement
func (r *registerTester) goMethod() (r2 *registerTester) {
	r2 = r
	r.wg.Add(1)
	return
}

// goFunction fixture is the function launching a new goroutine
//   - goFunction must have Register and NewCodeLocation on same line
//   - tt.goFunction is the goFunction of a goroutine
func (t *registerTester) goFunction() { t.g.Register(t.label); t.h(pruntime.NewCodeLocation(0)) }

// h stores register values and ends the goroutine
func (t *registerTester) h(g0RegisterLine *pruntime.CodeLocation) {
	defer t.wg.Done()

	t.cL = g0RegisterLine
	t.goroutineThreadID = goid.GoID()
}

// result awaits gorotuine eexits and returns actual values
func (r *registerTester) result() (cL *pruntime.CodeLocation, goroutineThreadID parl.ThreadID) {
	r.wg.Wait()
	cL = r.cL
	goroutineThreadID = r.goroutineThreadID
	return
}
