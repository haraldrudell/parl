/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"errors"

	"github.com/haraldrudell/parl/perrors"
)

// ErrNotCancelContext indicates that InvokeCancel was provided a context chain without NewCancelContext
var ErrNotCancelContext = errors.New("context chain does not have CancelContext")

type cancelContextKey string // type to use for context keys

var cancelKey cancelContextKey // key for context value

// NewCancelContext creates a context that can be provided to InvokeCancel
// the return value encapsulates a cancel function
func NewCancelContext(ctx context.Context) (cancelCtx context.Context) {
	return NewCancelContextFunc(context.WithCancel(ctx))
}

// NewCancelContextFunc creates a context invoking the provided cancel function on InvokeCancel
func NewCancelContextFunc(ctx context.Context, cancel context.CancelFunc) (cancelCtx context.Context) {
	return context.WithValue(ctx, cancelKey, cancel)
}

// InvokeCancel finds the cancel method and invokes it
func InvokeCancel(ctx context.Context) {
	cancel, ok := ctx.Value(cancelKey).(context.CancelFunc)
	if !ok {
		panic(perrors.Errorf("%v", ErrNotCancelContext))
	}
	cancel()
}

// CancelOnError invokes InvokeCancel if errp has an error
func CancelOnError(errp *error, ctx context.Context) {
	if errp == nil || *errp == nil {
		return // there was no error
	}
	InvokeCancel(ctx)
}
