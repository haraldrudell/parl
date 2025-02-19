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
	// error channel of initialized buffered length
	goResultChan
	// the cumulatice of all add provided to Remaining
	adds cyclebreaker.Atomic64[int]
	// the dimentioned size less
	remaining cyclebreaker.Atomic64[int]
	// true if any thread failed or SetIsError was invoked
	isError atomic.Bool
}

// NewGoResult2 also has isError
func newGoResultStruct(ch goResultChan) (goResult *goResultStruct) {
	g := goResultStruct{goResultChan: ch}
	g.remaining.Store(cap(ch))
	return &g
}

// ReceiveError is a deferrable function receiving error values from goroutines
//   - n is number of goroutines to wait for, default 1
//   - errp may be nil
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
	var remainingErrors int
	if len(n) > 0 {
		remainingErrors = n[0]
	} else {
		remainingErrors = int(g.remaining.Load())
	}

	// await goroutine results
	for ; remainingErrors > 0; remainingErrors-- {

		// blocks here
		//	- wait for a result from a goroutine
		var e = <-g.goResultChan
		if g.remaining.Add(-1) == 0 {
			break // end of configured errors
		} else if e == nil {
			continue // good return: ignore
		}

		// goroutine exited with error
		if !g.isError.Load() {
			g.isError.Store(true)
		}
		// ensure e has stack
		e = perrors.Stack(e)
		// build error list
		err = perrors.AppendError(err, e)
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

// Remaining returns the number of goroutines that have yet to exit
func (g *goResultStruct) Remaining(add ...int) (adds, remaining int) {
	var didAdd bool
	if len(add) > 0 {
		if a := add[0]; a != 0 {
			adds = g.adds.Add(a)
			didAdd = true
		}
	}
	if !didAdd {
		adds = g.adds.Load()
	}
	remaining = g.remaining.Load()

	return
}

func (g *goResultStruct) String() (s string) {
	return fmt.Sprintf("goResult_remain:%d_ch:%d(%d)_isError:%t",
		g.remaining.Load(),
		len(g.goResultChan), cap(g.goResultChan),
		g.IsError(),
	)
}
