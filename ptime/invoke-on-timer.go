/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"time"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors"
)

const (
	defaultInvokeOnTimerDuration = 100 * time.Millisecond
)

// InvokeOnTimer implements a timeout action
type InvokeOnTimer struct {
	awaitable cyclebreaker.Awaitable
}

// NewInvokeOnTimer panics or invokes elapsedFunc if not reset within d
//   - default d 100 ms
//   - on timer expiry:
//   - — if errPanic is non-nil, process-terminating panic
//   - — if elapsedFunc is non-nil, it is invoked.
//     Panics in elapsedFunc are recovered
//   - errPanic and elapsedFunc cannot both be nil
//   - each NewInvokeOnTimer launches a separate, blocked thread
//   - if errCh is present it should be a buffered channel receving
//     a value on thread-exit that may be a panic from the thread.
//     the thread should not have any panics but a panic may occur in
//     elapsedFunc.
//   - if errCh is missing or nil, panic is echoed to stderr
//   - resources are released on:
//   - — Stop invocation
//   - — timer firing
func NewInvokeOnTimer(d time.Duration, errPanic error, elapsedFunc func(), errorSink ...cyclebreaker.ErrorSink1) (timer *InvokeOnTimer) {
	if errPanic == nil && elapsedFunc == nil {
		cyclebreaker.NilError("errPanic and elapsedFunc")
	}
	if d <= 0 {
		d = defaultInvokeOnTimerDuration
	}
	var errorSink0 cyclebreaker.ErrorSink1
	if len(errorSink) > 0 {
		errorSink0 = errorSink[0]
	}

	var t = InvokeOnTimer{}
	go invokeOnTimerThread(
		time.NewTimer(d), t.awaitable.Ch(),
		errPanic, elapsedFunc,
		errorSink0,
	)
	return &t
}

// Stop cancels the timer and releases resources
func (t *InvokeOnTimer) Stop() { t.awaitable.Close() }

// the thread handling timer expiry
func invokeOnTimerThread(
	timer *time.Timer, c cyclebreaker.AwaitableCh,
	errPanic error, elapsedFunc func(),
	errorSink cyclebreaker.ErrorSink1,
) {
	var err error
	if errorSink != nil {
		errorSink.AddError(err)
	} else {
		defer infallible(&err)
	}
	// errPanic non-nil causes process termination by panic outside of RecoverErr
	if errPanic != nil {
		defer processTermination(&errPanic)
	}
	defer cyclebreaker.RecoverErr(func() cyclebreaker.DA { return cyclebreaker.A() }, &err)
	defer timer.Stop()

	select {
	case <-timer.C:
		if errPanic != nil {
			return // terminate the process
		}
		elapsedFunc() // panic will be recovered
	case <-c:
		errPanic = nil
	}
}

func infallible(errp *error) {
	var err = *errp
	if err == nil {
		return
	}
	cyclebreaker.Infallible.AddError(err)
}

func processTermination(errp *error) {
	var err = *errp
	if err == nil {
		return
	}
	// if the err has stack, ensure it is printed
	if perrors.HasStack(err) {
		cyclebreaker.Log(perrors.Long(err))
	}
	panic(err)
}
