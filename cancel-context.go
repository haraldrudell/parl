/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"errors"
	"sync/atomic"

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

type cancel1Notifier string

var notifier1Key cancel1Notifier

type NotifierFunc func(slice pruntime.StackSlice)

type atomicList *atomic.Pointer[[]NotifierFunc]

// AddNotifier adds a function that is invoked when any context is canceled
func AddNotifier(ctx context.Context, notifier NotifierFunc) (ctx2 context.Context) {
	return addNotifier(true, ctx, notifier)
}

// AddNotifier1 adds a function that is invoked when a child context is canceled
func AddNotifier1(ctx context.Context, notifier NotifierFunc) (ctx2 context.Context) {
	return addNotifier(false, ctx, notifier)
}

func addNotifier(allCancels bool, ctx context.Context, notifier NotifierFunc) (
	ctx2 context.Context) {
	if ctx == nil {
		panic(perrors.NewPF("ctx cannot be nil"))
	} else if notifier == nil {
		panic(perrors.NewPF("notifier cannot be nil"))
	}
	if !allCancels {
		return context.WithValue(ctx, notifier1Key, notifier)
	}
	atomp, ok := ctx.Value(notifierKey).(atomicList)
	if ok {
		for {
			var notifiersp = (*atomp).Load()
			var notifiers = append(*notifiersp, notifier)
			if (*atomp).CompareAndSwap(notifiersp, &notifiers) {
				return ctx
			}
		}
	}
	var notifiers = []NotifierFunc{notifier}
	var atomSlice atomic.Pointer[[]NotifierFunc]
	atomSlice.Store(&notifiers)
	atomp = &atomSlice
	return context.WithValue(ctx, notifierKey, atomp)
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

	// retrieve cancel function from the nearest context.valueCtx
	var cancel context.CancelFunc
	var ok bool
	if cancel, ok = ctx.Value(cancelKey).(context.CancelFunc); !ok {
		panic(perrors.Errorf("%v", ErrNotCancelContext))
	}
	// invoke the function canceling the context.valueCtx parent that is context.cancelCtx
	//	- and all its child contexts
	cancel()

	// fetch the nearest notify1 function
	var notifier, _ = ctx.Value(notifier1Key).(NotifierFunc)

	// fetch any notifyall list
	var notifiers []NotifierFunc
	if atomp, ok := ctx.Value(notifierKey).(atomicList); ok {
		notifiers = *(*atomp).Load()
	}

	if notifier == nil && len(notifiers) == 0 {
		return // no notifiers return
	}

	// stack trace for notifiers
	var cl = pruntime.NewStackSlice(cancelNotifierFrames)

	// invoke all notifier functions
	if notifier != nil {
		notifier(cl)
	}
	for _, n := range notifiers {
		n(cl)
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
