/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"math"
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
)

// when when GoResult.g is has multiple values
type goResultStruct struct {
	goResultChan
	remaining atomic.Uint64
	isError   atomic.Bool
}

// NewGoResult2 also has isError
func newGoResultStruct(ch goResultChan) (goResult *goResultStruct) {
	g := goResultStruct{goResultChan: ch}
	g.remaining.Store(uint64(cap(ch)))
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
		if g.remaining.Add(math.MaxUint64) == 0 {
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

func (g *goResultStruct) SetIsError() {
	if g.isError.Load() {
		return
	}
	g.isError.Store(true)
}

// IsError returns if any goroutine has returned an error
func (g *goResultStruct) IsError() (isError bool) { return g.isError.Load() }

// Remaining returns the number of goroutines that have yet to exit
func (g *goResultStruct) Remaining() (remaining int) { return int(g.remaining.Load()) }
