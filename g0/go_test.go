/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"errors"
	"testing"

	"github.com/haraldrudell/parl"
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
