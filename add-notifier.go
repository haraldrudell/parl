/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	// counts [parl.handleContextNotify] and [parl.invokeCancel]
	cancelNotifierFrames = 2
)

// notifier1Key is the context value key for child-context notifiers
var notifier1Key cancelContextKey = "notifyChild"

// a NotifierFunc receives a stack trace of function causing cancel
type NotifierFunc func(slice pruntime.StackSlice)

// notifierKey is the context value key for all-context notifiers
var notifierKey cancelContextKey = "notifyAll"

// atomicList is the type of context-value stored for all-cancel notifers
type atomicList *atomic.Pointer[[]NotifierFunc]

// AddNotifier adds a function that is invoked when any context is canceled
//   - AddNotifier is typically invoked on the root context
//   - any InvokeCancel in the context tree below the top AddNotifier
//     invocation causes notification
//   - invocation is immediately after context cancel completes
//   - implemented by inserting a thread-safe slice value into the context chain
//   - notifier receives a stack trace of the cancel invocation,
//     typically beginning with [parl.InvokeCancel]
//   - notifier should be thread-safe and not long running
//   - typical usage is debug of unexpected context cancel
func AddNotifier(ctx context.Context, notifier NotifierFunc) (ctx2 context.Context) {
	return addNotifier(true, ctx, notifier)
}

// AddNotifier1 adds a function that is invoked when a child context is canceled
//   - child contexts with their own AddNotifier1 are not detected
//   - invocation is immediately after context cancel completes
//   - implemented by inserting a value into the context chain
//   - notifier receives a stack trace of the cancel invocation,
//     typically beginning with [parl.InvokeCancel]
//   - notifier should be thread-safe and not long running
//   - typical usage is debug of unexpected context cancel
func AddNotifier1(ctx context.Context, notifier NotifierFunc) (ctx2 context.Context) {
	return addNotifier(false, ctx, notifier)
}

// addNotifier adds a notify context-value
func addNotifier(allCancels bool, ctx context.Context, notifier NotifierFunc) (
	ctx2 context.Context) {
	if ctx == nil {
		panic(perrors.NewPF("ctx cannot be nil"))
	} else if notifier == nil {
		panic(perrors.NewPF("notifier cannot be nil"))
	}

	// case for only child contexts
	if !allCancels {
		return context.WithValue(ctx, notifier1Key, notifier)
	}

	// append to static notifier list

	// if this context-chain has static notifier, append to it
	var atomp atomicList
	var ok bool
	atomp, ok = ctx.Value(notifierKey).(atomicList)
	if ok {
		for {
			var notifiersp = (*atomp).Load()
			var notifiers = append(*notifiersp, notifier)
			if (*atomp).CompareAndSwap(notifiersp, &notifiers) {
				return ctx // appended: done
			}
		}
	}

	// create a new list
	var atomSlice atomic.Pointer[[]NotifierFunc]
	atomSlice.Store(&[]NotifierFunc{notifier})

	// insert list pointer into context chain
	atomp = &atomSlice
	return context.WithValue(ctx, notifierKey, atomp)
}

// handleContextNotify is invoked for all CancelContext cancel invocations
func handleContextNotify(ctx context.Context) {
	// fetch the nearest notify1 function
	//	- notify1 are created by [parl.AddNotifier1] and are notified of
	//		cancellation of a child context
	var notifier, _ = ctx.Value(notifier1Key).(NotifierFunc)

	// fetch any notifyall list
	var notifiers []NotifierFunc
	if atomp, ok := ctx.Value(notifierKey).(atomicList); ok {
		notifiers = *(*atomp).Load()
	}

	if notifier == nil && len(notifiers) == 0 {
		return // no notifiers return
	}

	// stack trace for notifiers: expensive
	var cl = pruntime.NewStackSlice(cancelNotifierFrames)

	// invoke all notifier functions
	if notifier != nil {
		notifier(cl)
	}
	for _, n := range notifiers {
		n(cl)
	}
}
