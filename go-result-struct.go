/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"
	"sync/atomic"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
)

// goResultStruct is feature-rich [GoResult] implementation: struct with channel
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

// NewGoResult2 returns feature-rich [GoResult] implementation
func newGoResultStruct(ch goResultChan) (goResult *goResultStruct) {
	return &goResultStruct{goResultChan: ch}
}

// done delegates to goResultChan and counts invocations
func (g *goResultStruct) done(err error) {
	g.goResultChan.done(err)
	g.sendCount.Add(1)
}

// ch obtains the error providing channel
func (g *goResultStruct) ch() (ch <-chan error) { return g.goResultChan }

// count returns result status
func (g *goResultStruct) count() (available, stillRunning int) {
	available, stillRunning = len(g.goResultChan), g.adds.Load()
	if stillRunning == 0 {
		stillRunning = cap(g.goResultChan)
	}
	stillRunning -= g.sendCount.Load()
	if stillRunning < 0 {
		stillRunning = 0
	}

	return
}

// SetIsError sets the error flag regardless if any thread failed
func (g *goResultStruct) doError(setError bool) (isError bool) {

	// handle setError true
	if setError {
		isError = true
		if g.isError.Load() {
			return
		}
		g.isError.Store(true)
		return
	}

	isError = g.isError.Load()

	return
}

// Remaining returns the number of goroutines that should be awaited
//   - add: optional add for count-based number of created goroutines
//   - adds: the cumulative number of add values provided
//   - — adds allow for not waiting on goroutines that were never created
//   - if adds is zero, ie. no add was ever provided, adds is the dimensioned
//     capacity provided to the new-function
func (g *goResultStruct) remaining(addValue int) (adds int) {

	// do any add
	if addValue > 0 {
		adds = g.adds.Add(addValue)
		return
	}

	// get adds
	if adds = g.adds.Load(); adds > 0 {
		return
	}

	// use capacity if no adds
	adds = cap(g.goResultChan)

	return
}

// “goResult_adds:2_sends:1_ch:0(2)_isError:false”
func (g *goResultStruct) String() (s string) {
	return fmt.Sprintf("goResult_adds:%d_sends:%d_ch:%d(%d)_isError:%t",
		g.adds.Load(), g.sendCount.Load(),
		len(g.goResultChan), cap(g.goResultChan),
		g.isError.Load(),
	)
}
