/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"
	"sync/atomic"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors"
)

// when when GoResult.g is has multiple values
type goResultStruct struct {
	// error channel with capacity from new-function
	//	- has methods: Count() IsError() ReceiveError() Remaining()
	//		SendError() SetIsError() String()
	goResultChan
	// the cumulative of all add provided to Remaining
	//	- zero if no add were ever invoked
	adds cyclebreaker.Atomic64[int]
	// the number of SendError invocations
	sendCount cyclebreaker.Atomic64[int]
	// true if any thread failed or SetIsError was invoked
	isError atomic.Bool
}

// NewGoResult2 also has isError
func newGoResultStruct(ch goResultChan) (goResult *goResultStruct) {
	return &goResultStruct{goResultChan: ch}
}

func (g *goResultStruct) SendError(errp *error) {
	g.goResultChan.SendError(errp)
	g.sendCount.Add(1)
}

// ReceiveError is a deferrable function receiving error values from goroutines
//   - n: number of goroutines to wait for
//   - n missing: waits for adds if non-zero, otherwise new-function capacity
//   - errp: may be nil
//   - ReceiveError makes a goroutine:
//   - — awaitable and
//   - — able to return a fatal error
//   - — other needs of a goroutine is to initiate and detect cancel and
//     submit non-fatal errors
//   - GoRoutine should have enough capacity for all its goroutines
//   - — this prevents goroutines from blocking in channel send
//   - ReceiveError only panics from structural coding problems
//   - deferrable thread-safe
func (g *goResultStruct) ReceiveError(errp *error, n ...int) (err error) {

	// how many results to wait for
	var remainingErrors int
	if len(n) > 0 {
		remainingErrors = n[0]
	} else if remainingErrors = g.adds.Load(); remainingErrors == 0 {
		remainingErrors = cap(g.goResultChan)
	}

	// await goroutine results
	for ; remainingErrors > 0; remainingErrors-- {

		// blocks here
		//	- wait for a result from a goroutine
		var e = <-g.goResultChan
		if e == nil {
			continue // good return: ignore
		}
		// a goroutune exited with error

		//	flag error state
		if !g.isError.Load() {
			g.isError.Store(true)
		}

		// append to error list
		//	- // ensure e has stack
		err = perrors.AppendError(err, perrors.Stack(e))
	}

	// final action: update errp if present
	if err != nil && errp != nil {
		*errp = perrors.AppendError(*errp, err)
	}

	return
}

// SetIsError sets the error flag regardless if any thread failed
func (g *goResultStruct) SetIsError() {
	if g.isError.Load() {
		return
	}
	g.isError.Store(true)
}

// IsError returns if any goroutine has returned an error
func (g *goResultStruct) IsError() (isError bool) { return g.isError.Load() }

//   - available: the number of results that can be currently collected.
//     That is len of the result channel, ie.
//     SendError invocations yet to be collected by ReceiveError
//   - stillRunning [NewGoResult2] only: the number of created goroutines
//     yet to invoke SendError.
//     That is cumulative adds less SendError invocations.
//     If cumulative adds is zero, the dimensioned capacity provided to
//     new-function less SendError invocations
//   - Thread-safe
func (g *goResultStruct) Count() (available, stillRunning int) {
	available, stillRunning = len(g.goResultChan), g.adds.Load()
	if stillRunning == 0 {
		stillRunning = cap(g.goResultChan)
	}
	stillRunning -= g.sendCount.Load()

	return
}

// Remaining returns the number of goroutines that should be awaited
//   - add: optional add for count-based number of created goroutines
//   - adds: the cumulative number of add values provided
//   - — adds allow for not waiting on goroutines that were never created
//   - if adds is zero, ie. no add was ever provided, adds is the dimensioned
//     capacity provided to the new-function
func (g *goResultStruct) Remaining(add ...int) (adds int) {

	// do any add
	var a int
	if len(add) > 0 {
		if a = add[0]; a != 0 {
			adds = g.adds.Add(a)
		}
	}

	// ensure adds is read
	if a == 0 {
		adds = g.adds.Load()
	}
	// adds is g.adds

	// use capacity if no adds
	if adds == 0 {
		adds = cap(g.goResultChan)
	}

	return
}

func (g *goResultStruct) String() (s string) {
	return fmt.Sprintf("goResult_adds:%d_sends:%d_ch:%d(%d)_isError:%t",
		g.adds.Load(), g.sendCount.Load(),
		len(g.goResultChan), cap(g.goResultChan),
		g.IsError(),
	)
}
