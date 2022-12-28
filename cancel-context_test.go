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
