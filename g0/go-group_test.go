/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pruntime"
)

func TestGoGroup(t *testing.T) {
	var messageBad = "bad"
	var errBad = errors.New(messageBad)
	var shortTime = time.Millisecond

	var g parl.Go
	var goError parl.GoError
	var ok, isClosed, allowTermination, didReceive bool
	// var count, fatals int
	var subGo parl.SubGo
	var subGroup parl.SubGroup
	var ctx, ctxAct context.Context
	var cancelFunc context.CancelFunc
	// var threads []parl.ThreadData
	// var onFirstFatal = func(goGen parl.GoGen) { fatals++ }
	// var expectG0ID uint64
	var err error
	var noError *error
	var errCh <-chan parl.GoError
	// var didComplete atomic.Bool
	var isReady, isDone parl.WaitGroupCh
	var timer *time.Timer

	// Go() SubGo() SubGroup() Ch() Wait() EnableTermination()
	// IsEnableTermination() Cancel() Context() Threads() NamedThreads()
	// SetDebug()
	var goGroup parl.GoGroup
	var goGroupImpl *GoGroup
	var reset = func(ctx ...context.Context) {
		var parentContext context.Context
		if len(ctx) > 0 {
			parentContext = ctx[0]
		} else {
			parentContext = context.Background()
		}
		goGroup = NewGoGroup(parentContext)
		goGroupImpl = goGroup.(*GoGroup)
	}

	// NewGoGroup should not be canceled
	reset()
	isClosed = goGroupImpl.endCh.IsClosed()
	if isClosed {
		t.Error("NewGoGroup isClosed true")
	}

	// GoGroup should terminate when its last thread exits
	// Go should return parl.Go
	reset()
	g = goGroup.Go()
	g.Done(noError)
	isClosed = goGroupImpl.endCh.IsClosed()
	if !isClosed {
		t.Error("NewGoGroup last exit does not terminate")
	}

	// EnableTermination should be true
	reset()
	allowTermination = goGroup.EnableTermination()
	if !allowTermination {
		t.Error("EnableTermination false")
	}

	// EnableTermination(parl.PreventTermination) should be false
	reset()
	allowTermination = goGroup.EnableTermination(parl.PreventTermination)
	if allowTermination {
		t.Error("EnableTermination true")
	}

	// EnableTermination(parl.AllowTermination) should terminate an empty ThreadGroup
	reset()
	allowTermination = goGroup.EnableTermination(parl.AllowTermination)
	_ = allowTermination
	isClosed = goGroupImpl.endCh.IsClosed()
	if !isClosed {
		t.Error("EnableTermination true does not terminate")
	}

	// EnableTermination(parl.PreventTermination) should prevent termination
	reset()
	allowTermination = goGroup.EnableTermination(parl.PreventTermination)
	_ = allowTermination
	g = goGroup.Go()
	g.Done(noError)
	isClosed = goGroupImpl.endCh.IsClosed()
	if isClosed {
		t.Error("EnableTermination(parl.PreventTermination) does not prevent termination")
	}
	allowTermination = goGroup.EnableTermination(parl.AllowTermination)
	if !allowTermination {
		t.Error("EnableTermination false")
	}
	isClosed = goGroupImpl.endCh.IsClosed()
	if !isClosed {
		t.Error("EnableTermination(parl.AllowTermination) did not terminate")
	}

	// Context should return a context that is different from parent context
	ctx = context.Background()
	reset(ctx)
	ctxAct = goGroup.Context()
	if ctxAct == ctx {
		t.Error("Context is parent context")
	}

	// Context should return a context that is canceled by parent context
	ctx, cancelFunc = context.WithCancel(context.Background())
	reset(ctx)
	ctxAct = goGroup.Context()
	if ctxAct.Err() != nil {
		t.Error("Context is canceled")
	}
	cancelFunc()
	err = ctxAct.Err()
	if !errors.Is(err, context.Canceled) {
		t.Error("Context not canceled by parent")
	}

	// Cancel should cancel Context
	reset()
	ctxAct = goGroup.Context()
	if ctxAct.Err() != nil {
		t.Error("Context is canceled")
	}
	goGroup.Cancel()
	err = ctxAct.Err()
	if !errors.Is(err, context.Canceled) {
		t.Error("Cancel did not cancel Context")
	}

	// Cancel should cancel Go Context
	reset()
	g = goGroup.Go()
	goGroup.Cancel()
	err = g.Context().Err()
	if !errors.Is(err, context.Canceled) {
		t.Error("Cancel did not cancel Go Context")
	}

	// Cancel should cancel SubGo Context
	//	- SubGo
	reset()
	subGo = goGroup.SubGo()
	goGroup.Cancel()
	err = subGo.Context().Err()
	if !errors.Is(err, context.Canceled) {
		t.Error("Cancel did not cancel SubGo Context")
	}

	// Cancel should cancel SubGroup Context
	//	-SubGroup
	reset()
	subGroup = goGroup.SubGroup()
	goGroup.Cancel()
	err = subGroup.Context().Err()
	if !errors.Is(err, context.Canceled) {
		t.Error("Cancel did not cancel subGroup Context")
	}

	// Ch should send errors
	reset()
	errCh = goGroup.Ch()
	g = goGroup.Go()
	g.AddError(errBad)
	goError = <-errCh
	if !errors.Is(goError.Err(), errBad) {
		t.Error("Ch not sending errors")
	}

	// Ch should close on termination
	reset()
	goGroup.EnableTermination(parl.AllowTermination)
	select {
	case goError, ok = <-goGroup.Ch():
		didReceive = true
	default:
		didReceive = false
	}
	_ = goError
	if !didReceive || ok {
		t.Error("Ch did not close on termination")
	}

	// Wait should wait until GoGroup terminates
	reset()
	isReady.Reset().Add(1)
	isDone.Reset().Add(1)
	go waiter(goGroup, &isReady, &isDone)
	isReady.Wait()
	if isDone.IsZero() {
		t.Error("Wait completed prematurely")
	}
	goGroup.EnableTermination(parl.AllowTermination)
	// there is a race condition with waiter function
	//	- waiter needs to detect that the channel closed and
	//		trigger isDone
	//	- Wait enough here, shortTime
	timer = time.NewTimer(shortTime)
	select {
	case <-isDone.Ch():
	case <-timer.C:
	}
	if !isDone.IsZero() {
		t.Error("Wait did not complete on termination")
	}

	// methods to test below here:
	// Threads() NamedThreads()
	// SetDebug()
	//	- first fatal feature
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
	// threadGroup.(*GoGroup).addNotifier(func(slice pruntime.StackSlice) {
	// 	t.Logf("CANCEL %s %s", GoChain(threadGroup), slice)
	// })
	var subGroup = threadGroup.SubGroup()
	// subGroup.(*GoGroup).addNotifier(func(slice pruntime.StackSlice) {
	// 	t.Logf("CANCEL %s %s", GoChain(subGroup), slice)
	// })
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

func GoChain(g parl.GoGen) (s string) {
	for {
		var s0 = GoNo(g)
		if s == "" {
			s = s0
		} else {
			s += "—" + s0
		}
		if g == nil {
			return
		} else if g = Parent(g); g == nil {
			return
		}
	}
}

func Parent(g parl.GoGen) (parent parl.GoGen) {
	switch g := g.(type) {
	case *Go:
		parent = g.goParent.(parl.GoGen)
	case *GoGroup:
		if p := g.parent; p != nil {
			parent = p.(parl.GoGen)
		}
	}
	return
}

func GoNo(g parl.GoGen) (goNo string) {
	switch g1 := g.(type) {
	case *Go:
		goNo = "Go" + g1.id.String() + ":" + g1.GoID().String()
	case *GoGroup:
		if !g1.hasErrorChannel {
			goNo = "SubGo"
		} else if g1.parent != nil {
			goNo = "SubGroup"
		} else {
			goNo = "GoGroup"
		}
		goNo += g1.id.String()
	case nil:
		goNo = "nil"
	default:
		goNo = fmt.Sprintf("?type:%T", g)
	}
	return
}

// waiter tests GoGroup.Wait()
func waiter(
	goGroup parl.GoGroup,
	isReady, isDone parl.Doneable,
) {
	defer isDone.Done()

	isReady.Done()
	goGroup.Wait()
}
