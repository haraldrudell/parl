/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"sync/atomic"

	"github.com/haraldrudell/parl/pruntime"
)

const (
	// counts [parl.handleContextNotify] and [parl.invokeCancel]
	cancelNotifierFrames = 2
)

// a NotifierFunc receives a stack trace of function causing cancel
//   - typically stack trace begins with [parl.InvokeCancel]
type NotifierFunc func(slice pruntime.StackSlice)

// notifier1Key is the context value key for child-context notifiers
var notifier1Key cancelContextKey = "notifyChild"

// notifierAllKey is the context value key for all-context notifiers
var notifierAllKey cancelContextKey = "notifyAll"

// threadSafeList is a thread-safe slice for all-cancel notifers
type threadSafeList struct {
	// - atomic.Pointer.Load for thread-safe read
	// - CompareAndSwap of cloned list for thread-safe write
	notifiers atomic.Pointer[[]NotifierFunc]
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
	if ctx == nil {
		panic(NilError("ctx"))
	} else if notifier == nil {
		panic(NilError("notifier"))
	}
	return context.WithValue(ctx, notifier1Key, notifier)
}

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
	if ctx == nil {
		panic(NilError("ctx"))
	} else if notifier == nil {
		panic(NilError("notifier"))
	}

	// if this context-chain has static notifier, append to it
	if list, ok := ctx.Value(notifierAllKey).(*threadSafeList); ok { // ok only if non-nil
		for {
			// currentSlicep is read-only to be thread-safe
			var currentSlicep = list.notifiers.Load()
			// clone and append
			var newSlice = append(append([]NotifierFunc{}, *currentSlicep...), notifier)
			if list.notifiers.CompareAndSwap(currentSlicep, &newSlice) {
				ctx2 = ctx // appended: ctx does not change
				return     // append return
			}
		}
	}

	// create a new list
	var newList threadSafeList
	var newSlice = []NotifierFunc{notifier}
	newList.notifiers.Store(&newSlice)

	// insert list pointer into context chain
	ctx2 = context.WithValue(ctx, notifierAllKey, &newList)
	return // insert context value return, ctx2 new value
}

// handleContextNotify is invoked for all CancelContext cancel invocations
func handleContextNotify(ctx context.Context) {
	// fetch the nearest notify1 function
	//	- notify1 are created by [parl.AddNotifier1] and are notified of
	//		cancellation of a child context
	var notifier, _ = ctx.Value(notifier1Key).(NotifierFunc)

	// fetch any notifyall list
	var notifiers []NotifierFunc
	if list, ok := ctx.Value(notifierAllKey).(*threadSafeList); ok {
		notifiers = *list.notifiers.Load()
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
