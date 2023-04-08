/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"errors"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	cancelNotifierFrames = 1
)

// ErrNotCancelContext indicates that InvokeCancel was provided a context
// not a return value from NewCancelContext or NewCancelContextFunc.
//
// Tets for ErrNotCancelContext:
//
//	if errors.Is(err, parl.ErrNotCancelContext) …
var ErrNotCancelContext = errors.New("context chain does not have CancelContext")

// cancelContextKey is a unique named type used for access-function context.Context.Value.
// Because the cancelContextKey type is unique, there is no conflict with other values.
// cancelContextKey is used to store a context cancellation function as a context value
// so that Cancel can be invoked from a single context value.
type cancelContextKey string

// cancelKey is the value used for storing the cancel function with context.WIthValue.
var cancelKey cancelContextKey

type cancelNotifier string

var notifierKey cancelNotifier

func AddNotifier(ctx context.Context, notifierFn func(slice pruntime.StackSlice)) (
	ctx2 context.Context) {
	if ctx == nil {
		panic(perrors.NewPF("ctx cannot be nil"))
	}
	if notifierFn == nil {
		panic(perrors.NewPF("notifierFn cannot be nil"))
	}
	fnsAny := ctx.Value(notifierKey)
	fns, _ := fnsAny.([]func(slice pruntime.StackSlice))
	fns = append([]func(slice pruntime.StackSlice){notifierFn}, fns...)
	return context.WithValue(ctx, notifierKey, fns)
}

// NewCancelContext creates a context that can be provided to InvokeCancel.
// the return value encapsulates a cancel function.
//
//   - NewCancelContext is like [context.WithCancel] but with the CancelFunc embedded
//     instead, [InvokeCancel] is used  with cancelCtx as argument.
//
//     ctx := NewCancelContext(context.Background())
//     …
//     InvokeCancel(ctx)
func NewCancelContext(ctx context.Context) (cancelCtx context.Context) {
	return NewCancelContextFunc(context.WithCancel(ctx))
}

// NewCancelContextFunc stores the cancel function cancel in the context ctx.
// the returned context can be provided to InvokeCancel to cancel the context.
func NewCancelContextFunc(ctx context.Context, cancel context.CancelFunc) (cancelCtx context.Context) {
	return context.WithValue(ctx, cancelKey, cancel)
}

func HasCancel(ctx context.Context) (hasCancel bool) {
	if ctx == nil {
		return
	}
	cancel, _ := ctx.Value(cancelKey).(context.CancelFunc)
	return cancel != nil
}

// InvokeCancel finds the cancel method in the context chain and invokes it.
// ctx must have been returned by either NewCancelContext or NewCancelContextFunc.
//   - ctx nil is panic
//   - ctx not from NewCancelContext or NewCancelContextFunc is panic
//   - thread-safe, idempotent
func InvokeCancel(ctx context.Context) {
	invokeCancel(ctx)
}

func invokeCancel(ctx context.Context) {
	if ctx == nil {
		panic(perrors.NewPF("ctx cannot be nil"))
	}
	cancel, ok := ctx.Value(cancelKey).(context.CancelFunc)
	if !ok {
		panic(perrors.Errorf("%v", ErrNotCancelContext))
	}
	cancel()

	// invoke notifier
	fnsAny := ctx.Value(notifierKey)
	fns, _ := fnsAny.([]func(slice pruntime.StackSlice))
	if len(fns) == 0 {
		return
	}
	cl := pruntime.NewStackSlice(cancelNotifierFrames)
	for _, fn := range fns {
		fn(cl)
	}
}

// CancelOnError invokes InvokeCancel if errp has an error.
// CancelOnError is deferrable and thread-safe.
// ctx must have been returned by either NewCancelContext or NewCancelContextFunc.
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
