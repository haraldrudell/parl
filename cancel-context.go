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

// ErrNotCancelContext indicates that InvokeCancel was provided a context
// not a return value from NewCancelContext or NewCancelContextFunc.
//
// Test for ErrNotCancelContext:
//
//	if errors.Is(err, parl.ErrNotCancelContext) …
var ErrNotCancelContext = errors.New("context chain does not have CancelContext")

// cancelContextKey is a unique named for storing and retrieving cancelFunc
//   - used with [context.WithValue]
type cancelContextKey string

// cancelKey is a unique value for storing and retrieving cancelFunc
//   - used with [context.WithValue]
var cancelKey cancelContextKey = "parl.NewCancelContext"

// NewCancelContext creates a cancelable context without managing a CancelFunction value
//   - NewCancelContext is like [context.WithCancel] but has the CancelFunc embedded.
//   - after use, [InvokeCancel] must be invoked with cancelCtx as argument to
//     release resources
//   - —
//   - for unexported code in context package to work, a separate type cannot be used
//
// Usage:
//
//	ctx := NewCancelContext(context.Background())
//	…
//	InvokeCancel(ctx)
func NewCancelContext(ctx context.Context) (cancelCtx context.Context) {
	return NewCancelContextFunc(context.WithCancel(ctx))
}

// NewCancelContextFunc stores the cancel function cancel in the context ctx.
// the returned context can be provided to InvokeCancel to cancel the context.
func NewCancelContextFunc(ctx context.Context, cancel context.CancelFunc) (cancelCtx context.Context) {
	return context.WithValue(ctx, cancelKey, cancel)
}

// HasCancel return if ctx can be used with [parl.InvokeCancel]
//   - such contextx are returned by [parl.NewCancelContext]
func HasCancel(ctx context.Context) (hasCancel bool) {
	if ctx == nil {
		return
	}
	cancel, _ := ctx.Value(cancelKey).(context.CancelFunc)
	return cancel != nil
}

// InvokeCancel cancels the last CancelContext in ctx’ chain of contexts
//   - ctx must have been returned by either NewCancelContext or NewCancelContextFunc
//   - ctx nil is panic
//   - ctx not from NewCancelContext or NewCancelContextFunc is panic
//   - thread-safe, idempotent, deferrable
func InvokeCancel(ctx context.Context) {
	invokeCancel(ctx)
}

// invokeCancel is one extra stack frame from invoker
//   - invoked via:
//   - — [parl.InvokeCancel]
//   - — [parl.CancelOnError]
//   - — [parl.OnceWaiter.Cancel]
func invokeCancel(ctx context.Context) {
	if ctx == nil {
		panic(perrors.NewPF("ctx cannot be nil"))
	}

	// retrieve cancel function from the nearest context.valueCtx
	var cancel context.CancelFunc
	var ok bool
	if cancel, ok = ctx.Value(cancelKey).(context.CancelFunc); !ok {
		panic(perrors.ErrorfPF("%w", ErrNotCancelContext))
	}

	// invoke the function canceling the context.valueCtx parent that is context.cancelCtx
	//	- and all its child contexts
	cancel()

	handleContextNotify(ctx)
}

// CancelOnError invokes InvokeCancel if errp has an error.
//   - CancelOnError is deferrable and thread-safe.
//   - ctx must have been returned by either NewCancelContext or NewCancelContextFunc.
//   - errp == nil or *errp == nil means no error
//   - ctx nil is panic
//   - ctx not from NewCancelContext or NewCancelContextFunc is panic
//   - thread-safe, idempotent
func CancelOnError(errp *error, ctx context.Context) {
	if errp == nil || *errp == nil {
		return // there was no error
	}
	invokeCancel(ctx)
}
