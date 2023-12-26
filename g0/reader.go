/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// for [Reader]
var NoShouldTerm *atomic.Bool

// for [Reader]
var NoAddErr parl.AddError

// for [Reader]
var NoLog parl.PrintfFunc

// Reader thread reads the error channel of a [GoGroup] or [SubGroup]
//   - shouldTerminate is an optional pointer that an application’s most important goroutine sets to true
//     prior to exit, causing graceful shutdown
//   - addError receives fatal thread-exits.
//     If not present, those are logged.
//     typically [github.com/haraldrudell/parl/mains.Executable.AddErr]
//   - log outputs warnings and more, default [parl.Log] standard error
//   - g is from [parl.NewGoResult] or [parl.NewGoResult2] making Reader awaitable
//   - a GoGroup’s or SubGroup’s error channel is unbound buffer so Reader is only required for:
//   - — real-time warning output
//   - — terminating the process while additional goroutines are still running:
//   - — on fatal thread exit or
//   - — on exit of a primary goroutine
//   - —
//   - because reading of the threadgroup’s error channel must not stop,
//     it is done in this separate thread.
//   - reading continues until:
//   - — the threadGroup context is canceled by eg. [GoGroup.Cancel]
//   - — the last thread exits
//   - — a thread exits with error
//   - — on thread exit, shouldTerminate is true
//
// Usage:
//
//	func main() {
//	  var err error
//	  defer mains.MinimalRecovery(&err)
//	  var goGroup = g0.NewGoGroup(context.Background())
//	  defer goGroup.Wait()
//	  var goResult = parl.NewGoResult()
//	  defer goResult.ReceiveError(&err)
//	  go g0.Reader(g0.NoSholdTerm, g0.NoAddErr, g0.NoLog, goGroup, goResult)
//	  defer goGroup.Cancel()
//	  go someGoroutine(goGroup.Go())
func Reader(shouldTerminate *atomic.Bool, addError parl.AddError, log parl.PrintfFunc, goGroup parl.GoGroup, g parl.GoResult) {
	var err error
	defer g.SendError(&err)
	defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err)

	if log == nil {
		log = parl.Log
	}
	for goError := range goGroup.Ch() {

		// if not thread-exit, it is a warning
		//	- if panic: full stack trace
		//	- otherwise just error location
		if !goError.IsThreadExit() {
			log("Warning: " + goError.ErrString())
			continue // warning processed
		}

		// error from exiting goroutine
		var e = goError.Err()

		// no thread should return [context.Canceled]
		//	- on context cancel threads should exit silently
		//	- here is printed any thread ID returning context cancel
		if e != nil {
			var gotContextCancel string
			// error may be associated by using [perrors.AppendError]
			for _, anError := range perrors.ErrorList(e) {
				if errors.Is(anError, context.Canceled) {
					gotContextCancel = "context.Canceled"
					break
				}
			}
			if gotContextCancel != "" {
				var g = goError.Go()
				log("BAD: %s emitted by goroutine#%d func: %s trace:\n%s",
					gotContextCancel,
					g.GoID(),
					g.ThreadInfo().Func().Short(), // the function launching the goroutine
					perrors.Long(e),               // stack trace for the main error having context.Canceled
				)
			}
		}

		// fatal thread-exit shuts down the app
		if e != nil {
			if addError != nil {
				addError(e)
			} else {
				log("FATAL: " + goError.ErrString())
			}
			goGroup.Cancel()
			continue // fatal exit processed
		}

		// shouldTerminate is for apps that has a primary sub-thread that on exit
		// should shut down the app
		if t := shouldTerminate; t != nil && t.Load() && goGroup.Context().Err() == nil {
			goGroup.Cancel()
			// shouldTerminate processed
		}
	}

	// graceful threadGroup termination
	log("threadGroup ended")
}
