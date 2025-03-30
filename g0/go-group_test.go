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
	// isDebug is true if test is being debugged and may
	// be stopped long time at a breakpoint
	var isDebug bool //= true
	// timeout is timeout to wait for data
	//	- isDebug true: timeout zero: wait indefinitely while stopped in debugger
	//	- isDebug false: timeout 1ms so test does not hang indefinitely
	var timeout time.Duration
	if isDebug {
		t.Errorf("isDebug: timeouts are disabled")
	} else {
		timeout = time.Millisecond
	}
	const (
		// shortTime waits slightly to avoid race-condition errors
		shortTime = time.Millisecond
		// messageBad is error message fixture
		messageBad = "bad"
	)
	var (
		// errBad is error fixture
		errBad = errors.New(messageBad)
	)

	var (
		g                parl.Go
		goError          parl.GoError
		isClosed         bool
		allowTermination parl.GoTermination
		subGo            parl.SubGo
		subGroup         parl.SubGroup
		ctx, ctxAct      context.Context
		cancelFunc       context.CancelFunc
		err              error
		noError          *error
		isReady, isDone  parl.WaitGroupCh
		timer            *time.Timer
		goGroupImpl      *GoGroup
		isTimeout        timeoutFlag
		hasValue         bool
	)

	// Go SubGo SubGroup GoError Wait EnableTermination
	// Cancel Context Threads NamedThreads SetDebug
	var goGroup parl.GoGroup
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

	// Go from GoGroup good exit
	// Go()
	reset()
	// GoGroup should terminate when its last thread exits
	g = goGroup.Go()
	g.Done(noError)
	isClosed = goGroupImpl.endCh.IsClosed()
	if !isClosed {
		t.Error("NewGoGroup last exit does not terminate")
	}

	// EnableTermination should be true on new GoGroup
	reset()
	allowTermination = goGroup.EnableTermination()
	if allowTermination != parl.AllowTermination {
		t.Errorf("EnableTermination %s not %s", allowTermination, parl.AllowTermination)
	}

	// EnableTermination(parl.PreventTermination) should be false
	reset()
	allowTermination = goGroup.EnableTermination(parl.PreventTermination)
	if allowTermination != parl.PreventTermination {
		t.Errorf("EnableTermination %s not %s", allowTermination, parl.PreventTermination)
	}

	// EnableTermination(parl.AllowTermination)
	// should terminate an empty ThreadGroup
	reset()
	allowTermination = goGroup.EnableTermination(parl.AllowTermination)
	_ = allowTermination
	isClosed = goGroupImpl.endCh.IsClosed()
	if !isClosed {
		t.Error("EnableTermination true does not terminate")
	}

	// EnableTermination(parl.PreventTermination)
	// should prevent termination
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
	if allowTermination != parl.AllowTermination {
		t.Errorf("EnableTermination %s not %s", allowTermination, parl.AllowTermination)
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

	// GoError should send errors
	reset()
	goGroup.Go().AddError(errBad)
	goError, hasValue, isTimeout = awaitGoError(&goGroupImpl.goErrorStream, timeout)
	_ = hasValue
	if isTimeout == timeoutYES {
		t.Fatal("GoGroup error channel timeout")
	}
	if goError == nil {
		t.Fatal("goError nil")
	}
	if !errors.Is(goError.Err(), errBad) {
		t.Error("GoError sending bad errors")
	}

	// goErrorStream should close on termination
	//	- goErrorStream is not exposed by GoGroup
	//	- access the package-private goErrorStream using goGroupImpl
	reset()
	// AllowTermination for unused GoGroup causes it to terminate
	goGroup.EnableTermination(parl.AllowTermination)
	select {
	case <-goGroupImpl.goErrorStream.CloseCh():
	default:
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
	var (
		goGroup  parl.GoGroup
		subGo    parl.SubGo
		subGroup parl.SubGroup
		g        parl.Go
		cL       *pruntime.CodeLocation
	)

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
	g, cL = goGroup.Go(), pruntime.NewCodeLocation(0)
	if !strings.HasSuffix(g.String(), cL.Short()) {
		var _ = (&GoGroup{}).Go
		// Go.String: "subGroup#3_threads:0(0)_New:g0.TestGoGroup_Frames()-go-group_test.go:217" exp suffix: "g0.TestGoGroup_Frames()-go-group_test.go:223"
		t.Errorf("Go.String: %q exp suffix: %q", subGroup.String(), cL.Short())
	}
}

func TestSubGo(t *testing.T) {
	var isDebug bool //= true
	var timeout time.Duration
	if isDebug {
		t.Errorf("isDebug: timeouts are disabled")
	} else {
		timeout = time.Millisecond
	}
	var (
		err = errors.New("bad")
	)

	var (
		goGroup                parl.GoGroup
		goGroupImpl, subGoImpl *GoGroup
		goError, goError2      parl.GoError
		g                      parl.Go
		hasValue               bool
		isTimeout              timeoutFlag
	)

	// Go Subgo SubGroup Wait WaitCh EnableTermination
	// Cancel Context Threads NamedThreads SetDebug
	var subGo parl.SubGo
	var reset = func() {
		goGroup = NewGoGroup(context.Background())
		goGroupImpl = goGroup.(*GoGroup)
		subGo = goGroup.SubGo()
		subGoImpl = subGo.(*GoGroup)
	}

	// SubGo non-fatal error
	reset()
	// a non-fatal subGo error should be recevied on GoGroup error channel
	goError = NewGoError(err, parl.GeNonFatal, subGo.Go())
	subGoImpl.ConsumeError(goError)
	goError2, hasValue, isTimeout = awaitGoError(&goGroupImpl.goErrorStream, timeout)
	_ = hasValue
	if isTimeout == timeoutYES {
		t.Fatal("GoGroup error channel timeout")
	}
	if goError2 != goError {
		t.Errorf("bad non-fatal subgo error")
	}

	// SubGo fatal thread termination
	reset()
	// a SubGo fatal error should be recevied on GoGroup error channel
	g = subGo.Go()
	g.Done(&err)
	goError2, hasValue, isTimeout = awaitGoError(&goGroupImpl.goErrorStream, timeout)
	_ = hasValue
	if isTimeout == timeoutYES {
		t.Fatal("GoGroup error channel timeout")
	}
	if goError2 == nil {
		t.Fatal("goErorr nil")
	}
	if !errors.Is(goError2.Err(), err) {
		t.Error("bad fatal subgo error")
	}
	// subgo should terminate after its only thread exited
	if !subGoImpl.isEnd() {
		t.Error("subGo did not terminate")
	}
	// gogroup should end and closed its error channel
	//	- its only thread did exit
	select {
	case <-goGroupImpl.goErrorStream.CloseCh():
	default:
		t.Errorf("goGroup channel did not close: %s", goError2)
	}
	if !goGroupImpl.isEnd() {
		t.Error("goGroup did not terminate")
	}
}

func TestSubGroup(t *testing.T) {
	var isDebug bool //= true
	var timeout time.Duration
	if isDebug {
		t.Errorf("isDebug: timeouts are disabled")
	} else {
		timeout = time.Millisecond
	}
	var (
		err = errors.New("bad")
	)

	var (
		goGroup                       parl.GoGroup
		goGroupImpl, subGroupImpl     *GoGroup
		goError, goError2             parl.GoError
		g                             parl.Go
		hasValue                      bool
		isTimeout                     timeoutFlag
		goGroupErrors, subGroupErrors parl.Source1[parl.GoError]
	)

	// Go SubGo SubGroup GoError Wait WaitCh EnableTermination
	// Cancel Context Threads NamedThreads SetDebug FirstFatal
	var subGroup parl.SubGroup
	var reset = func() {
		goGroup = NewGoGroup(context.Background())
		goGroupImpl = goGroup.(*GoGroup)
		goGroupErrors = &goGroupImpl.goErrorStream
		subGroup = goGroup.SubGroup()
		subGroupImpl = subGroup.(*GoGroup)
		subGroupErrors = &subGroupImpl.goErrorStream
	}

	// A SubGroup non-fatal error should be read from GoGroup
	reset()
	goError = NewGoError(err, parl.GeNonFatal, subGroup.Go())
	subGroupImpl.ConsumeError(goError)
	goError2, hasValue, isTimeout = awaitGoError(goGroupErrors, timeout)
	// because goError2 is an error, fmt %s will print the error message
	_ = hasValue
	if isTimeout == timeoutYES {
		t.Fatal("GoGroup error channel timeout")
	}
	if goError2 != goError {
		t.Errorf("bad non-fatal subgroup error")
	}

	// fatal error by last thread of SubGroup
	//	- a thread exits with g0.Done having error
	//	- the subGroup hides the fatal error from the parent
	//	- the parent receives non-fatal GeLocalChan of the error and a GeExit with no error
	//	- subgroup emits fatal error on its error channel
	reset()
	// A SubGroup fatal error should be GeLocalChan GoError with GoGroup
	//	- this prevents fatal SubGroup terminations from terminating the GoGroup
	//	- while those terminations are printed as non-fatal errors
	g = subGroup.Go()
	// this done ends the last thread of SubGroup and goGroup
	//	- causing subGroup to end once its error stream is read to end
	//	- causing goGroup to end once its error streamis read to end
	g.Done(&err)
	// get the first of two errors from goGroup
	goError2, hasValue, isTimeout = awaitGoError(goGroupErrors, timeout)
	_ = hasValue
	if isTimeout == timeoutYES {
		t.Fatal("GoGroup error channel timeout")
	}
	if goError2 == nil {
		t.Fatal("GoGroup error streamclosed")
	}
	if !errors.Is(goError2.Err(), err) {
		t.Error("bad gogroup error")
	}
	if goError2.ErrContext() != parl.GeLocalChan {
		t.Errorf("bad gogroup error context: %s", goError2.ErrContext())
	}
	// The SubGroup fatal exit should addditionally be a GeExit good exit with GoGroup
	//	- this ensures that GoGroup’s thread count is accurate
	//	- get the last error from gogroup causing it to end
	goError2, hasValue, isTimeout = awaitGoError(goGroupErrors, timeout)
	_ = hasValue
	if isTimeout == timeoutYES {
		t.Fatal("GoGroup error channel timeout")
	}
	if goError2 == nil {
		t.Fatal("GoGroup error streamclosed")
	}
	if goError2.Err() != nil {
		t.Errorf("bad gogroup error: %s", goError2.String())
	}
	if goError2.ErrContext() != parl.GeExit {
		t.Errorf("bad gogroup error context: %s", goError2.ErrContext())
	}
	// The SubGroup fatal exit should be GeExit fatal error with SubGroup
	goError2, hasValue, isTimeout = awaitGoError(subGroupErrors, timeout)
	_ = hasValue
	if isTimeout == timeoutYES {
		t.Fatal("SubGroup error channel timeout")
	}
	if goError2 == nil {
		t.Fatal("SubGroup error streamclosed")
	}
	if goError2.ErrContext() != parl.GeExit {
		t.Errorf("bad subgroup error context: %s", goError2.ErrContext())
	}
	if !errors.Is(goError2.Err(), err) {
		t.Error("bad fatal subgroup error")
	}
	// subgroup should exit when it last thread exits
	select {
	case <-subGroupImpl.goErrorStream.CloseCh():
	default:
		t.Error("subGroup did not terminate")
	}
	// gogroup should exit when its last thread exits
	select {
	case <-goGroupImpl.goErrorStream.CloseCh():
	default:
		t.Errorf("goGroup did not terminate %s", goGroup)
	}
}

func TestCancel(t *testing.T) {
	var ctx = parl.AddNotifier(context.Background(), func(stack parl.Stack) {
		t.Logf("ALLCANCEL %s", stack)
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

func TestGoGroupTermination(t *testing.T) {
	var goGroup = NewGoGroup(context.Background())

	// an unused goGroup will only terminate after EnableTermination
	goGroup.EnableTermination(parl.AllowTermination)

	goGroup.Wait()
}

func TestSubGoTermination(t *testing.T) {
	const (
		timeout = time.Second
	)
	var (
		goGroup parl.GoGroup
		subGo   parl.SubGo
		timer   *time.Timer
	)

	// an unused subGo should terminate on EnableTermination AllowTermination
	goGroup = NewGoGroup(context.Background())
	subGo = goGroup.SubGo()
	subGo.EnableTermination(parl.AllowTermination)

	// subGo.Wait() should not block
	timer = time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-subGo.WaitCh():
		timer.Stop()
	case <-timer.C:
		t.Fatal("SubGo failed to terminate")
	}

	// CascadeTermination does this
	//goGroup.EnableTermination(parl.AllowTermination)

	// goGroup.Wait() should not block
	timer = time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-goGroup.WaitCh():
		timer.Stop()
	case <-timer.C:
		t.Fatal("GoGroup failed to terminate")
	}
}

func TestGoGroup2Termination(t *testing.T) {
	var goGroup = NewGoGroup(context.Background())
	var subGroup = goGroup.SubGroup()
	var subGo = subGroup.SubGo()

	// an unused subGo will only terminate after EnableTermination
	subGo.EnableTermination(parl.AllowTermination)

	//subGo.EnableTermination(parl.AllowTermination)
	subGo.Wait()

	subGroup.Wait()

	goGroup.Wait()
}

// waiter tests GoGroup.Wait()
func waiter(
	goGroup parl.GoGroup,
	isReady, isDone parl.DoneLegacy,
) {
	defer isDone.Done()

	isReady.Done()
	goGroup.Wait()
}

// awaitGoError awaits a single GoError optionally with timeout
//   - goErrors: the GoError stream source.
//     To await using select statement, an iterator does not work.
//     Must have the stream-source object
//   - timeout: 0 or timeout
//   - goError: any GoError value
//   - hasValue: true if a value was read
//   - isTimeout: true if receive timed out
func awaitGoError(goErrors parl.Source1[parl.GoError], timeout time.Duration) (goError parl.GoError, hasValue bool, isTimeout timeoutFlag) {

	// C is timeout channel or nil if no timeout
	//	- nil waits indefinitely
	var C <-chan time.Time
	if timeout > 0 {
		var timer = time.NewTimer(timeout)
		defer timer.Stop()
		C = timer.C
	}

	// await error or timeout
	select {
	case <-goErrors.DataWaitCh():
		goError, hasValue = goErrors.Get()
	case <-C:
		isTimeout = timeoutYES
	}

	return
}

const (
	// timeoutYES means timeout occured
	timeoutYES timeoutFlag = 1
)

// timeoutFlag is typesafe timeout indicator
//   - [timeoutYES] means timeout occured
type timeoutFlag int
