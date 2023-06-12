/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

func TestInvokeCancel(t *testing.T) {

	ctx := NewCancelContext(context.Background())

	InvokeCancel(ctx)

	if ctx.Err() == nil {
		t.Log("InvokeCancel failed")
	}

	var fInvoked bool
	f := func() {
		fInvoked = true
	}

	NewCancelContextFunc(context.Background(), f)

	if !fInvoked {
		t.Log("InvokeCancel Func failed")
	}
}

func TestCancelOnError(t *testing.T) {
	var err error

	CancelOnError(nil, nil)
	CancelOnError(&err, nil)
	ctx := NewCancelContext(context.Background())
	err = errors.New("x")
	CancelOnError(&err, ctx)
	if ctx.Err() == nil {
		t.Log("CancelOnError failed")
	}
}

func TestNewCancelContextNil(t *testing.T) {
	message := "from nil parent"

	var ctx context.Context
	var err error

	func() {
		defer func() {
			if v := recover(); v != nil {
				var ok bool
				if err, ok = v.(error); !ok {
					err = perrors.Errorf("panic-value not error; %T '%[1]v'", v)
				}
			}
		}()
		NewCancelContext(ctx)
	}()
	if err == nil || !strings.Contains(err.Error(), message) {
		t.Errorf("InvokeCancel bad error: %v exp %q", err, message)
	}
}

func TestInvokeCancelNil(t *testing.T) {
	messageCtxNil := "ctx cannot be nil"

	var ctx context.Context
	var err error

	func() {
		defer func() {
			if v := recover(); v != nil {
				var ok bool
				if err, ok = v.(error); !ok {
					err = perrors.Errorf("panic-value not error; %T '%[1]v'", v)
				}
			}
		}()
		InvokeCancel(ctx)
	}()
	if err == nil || !strings.HasSuffix(err.Error(), messageCtxNil) {
		t.Errorf("InvokeCancel bad error: '%v' exp %q", err, messageCtxNil)
	}
}

func TestInvokeCancelBad(t *testing.T) {
	message := "context chain does not have CancelContext"

	var ctx context.Context
	var err error

	ctx = context.Background()
	func() {
		defer func() {
			if v := recover(); v != nil {
				var ok bool
				if err, ok = v.(error); !ok {
					err = perrors.Errorf("panic-value not error; %T '%[1]v'", v)
				}
			}
		}()
		InvokeCancel(ctx)
	}()
	if err == nil || !strings.HasSuffix(err.Error(), message) {
		t.Errorf("InvokeCancel bad error: %v exp %q", err, message)
	}
}

func TestChildCancel(t *testing.T) {
	var v any
	var a func()
	var ok bool
	a, ok = v.(func())
	_ = a
	_ = ok
	var ctx0 = context.Background()
	var ctx1 = AddNotifier(ctx0, func(slice pruntime.StackSlice) {
		t.Log(slice)
	})
	var ctx2 = NewCancelContext(ctx1)
	var ctx3 = NewCancelContext(ctx2)
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

type NotifierCounter struct {
	count int
	t     *testing.T
}

func (c *NotifierCounter) Notifier(slice pruntime.StackSlice) {
	c.count++
	var t = c.t
	t.Logf("TRACE: %s", pruntime.NewStackSlice(0))
}

func TestCancels(t *testing.T) {
	var expCount = 3 // 2 notifierAll + notifier1

	var c = NotifierCounter{t: t}
	var anyValue any
	var ctx = NewCancelContext(AddNotifier(
		AddNotifier(
			AddNotifier1(
				AddNotifier1(context.Background(), c.Notifier),
				c.Notifier,
			),
			c.Notifier,
		),
		c.Notifier,
	))
	anyValue = ctx.Value(notifier1Key)
	if IsNil(anyValue) {
		t.Errorf("Notifier1 NIL")
	}
	t.Logf("ONE: %T", anyValue)
	anyValue = ctx.Value(notifierKey)
	if IsNil(anyValue) {
		t.Errorf("Notifier MANY NIL")
	}
	t.Logf("MANY: %T", anyValue)
	invokeCancel(ctx)
	if c.count != expCount {
		t.Errorf("count: %d exp %d", c.count, expCount)
	}
}
