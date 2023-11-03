/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/haraldrudell/parl/errorglue"
	"github.com/haraldrudell/parl/pruntime"
)

// tests:
// NewCancelContext
// NewCancelContextFunc
// InvokeCancel
// HasCancel
func TestInvokeCancel(t *testing.T) {
	// nil parent”
	var messageNilParent = "nil parent"
	// “ctx cannot be nil”
	var messageCtxNil = "ctx cannot be nil"

	var ctx context.Context
	var err error
	var isPanic bool

	// regular context
	ctx = context.Background()
	// HasCancel should be false
	if HasCancel(ctx) {
		t.Error("context HasCancel")
	}
	// InvokeCancel should panic
	isPanic, err = invokeInvokeCancel(ctx)
	if !isPanic || err == nil {
		t.Error("InvokeCancel no panic")
	}
	if !errors.Is(err, ErrNotCancelContext) {
		t.Errorf("err not ErrNotCancelContext %s “%+[1]v”",
			errorglue.DumpChain(err),
			err,
		)
	}

	// cancelContext
	ctx = NewCancelContext(ctx)
	// should not be canceled
	if ctx.Err() != nil {
		t.Error("context was canceled")
	}
	// HasCancel should be true
	if !HasCancel(ctx) {
		t.Error("context not HasCancel")
	}

	// InvokeCancel should cancel the context
	InvokeCancel(ctx)
	if ctx.Err() == nil {
		t.Error("InvokeCancel did not cancel")
	}

	ctx = nil
	_, isPanic, err = invokeNewCancelContext(ctx)
	if !isPanic {
		t.Error("NewCancelContext nil no panic")
	}
	if err == nil || !strings.Contains(err.Error(), messageNilParent) {
		t.Errorf("InvokeCancel bad error: %v exp %q", err, messageNilParent)
	}

	isPanic, err = invokeInvokeCancel(ctx)
	if !isPanic {
		t.Error("InvokeCancel nil no panic")
	}
	if err == nil || !strings.Contains(err.Error(), messageCtxNil) {
		t.Errorf("InvokeCancel bad error: '%v' exp %q", err, messageCtxNil)
	}
}

func TestCancelOnError(t *testing.T) {
	var err error

	// nil,nil should not panic
	CancelOnError(nil, nil)

	CancelOnError(&err, nil)
	ctx := NewCancelContext(context.Background())
	err = errors.New("x")
	CancelOnError(&err, ctx)
	if ctx.Err() == nil {
		t.Log("CancelOnError failed")
	}
}

func TestAfterFunc(t *testing.T) {
	var duration = time.Second

	var ctx = NewCancelContext(context.Background())
	var counter = newInvokeDetector()
	context.AfterFunc(ctx, counter.Func)

	// invokeCancel should start counter.Func in separate goroutine
	InvokeCancel(ctx)
	counter.waitForCh(duration)
	if c := counter.u32.Load(); c != 1 {
		t.Errorf("Bad number of invocations: %d exp 1", c)
	}
}

// invokeInvokeCancel calls InvokeCancel recovering panic
func invokeInvokeCancel(ctx context.Context) (isPanic bool, err error) {
	defer PanicToErr(&err, &isPanic)

	InvokeCancel(ctx)
	return
}

// invokeNewCancelContext invokes NewCancelContext recovering panic
func invokeNewCancelContext(ctx context.Context) (ctx2 context.Context, isPanic bool, err error) {
	defer PanicToErr(&err, &isPanic)

	ctx2 = NewCancelContext(ctx)

	return
}

// invokeDetector counts invocations
type invokeDetector struct {
	u32 atomic.Uint32
	ch  chan struct{}
}

// newInvokeDetector retruns an object counting invocations
func newInvokeDetector() (i *invokeDetector) { return &invokeDetector{ch: make(chan struct{})} }

func (i *invokeDetector) Func() {
	if i.u32.Add(1) == 1 {
		close(i.ch)
	}
}

// waitForCh wait up to d for first Func invocation
func (i *invokeDetector) waitForCh(d time.Duration) {
	var timer = time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-i.ch:
	case <-timer.C:
	}
}

func TestChildCancel(t *testing.T) {
	var v any
	var a func()
	var ok bool
	a, ok = v.(func())
	_ = a
	_ = ok

	// create ctx0…3
	var ctx0 = context.Background()
	var ctx1 = AddNotifier(ctx0, func(slice pruntime.StackSlice) {
		t.Log(slice)
	})
	var ctx2 = NewCancelContext(ctx1)
	var ctx3 = NewCancelContext(ctx2)

	// cancel ctx3 should cancel only ctx3
	invokeCancel(ctx3)
	if ctx3.Err() == nil {
		t.Error("ctx3 not canceled")
	}
	if ctx2.Err() != nil {
		t.Error("ctx2 canceled")
	}
	if ctx1.Err() != nil {
		t.Error("ctx1 canceled")
	}
	if ctx0.Err() != nil {
		t.Error("ctx0 canceled")
	}
}
