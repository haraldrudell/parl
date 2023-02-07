/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"time"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/ptime"
	"github.com/haraldrudell/parl/sets"
)

var DidOnceReturn AtomicCounter
var NewCalculationReturn AtomicCounter
var AfterReturn AtomicCounter
var WinnerReturn AtomicCounter
var CancelReturn AtomicCounter
var Arrival ptime.EpochValue
var Result ptime.EpochValue
var Calc AtomicCounter

const (
	// WinOrWaiterAnyValue causes a thread to accept any calculated value
	WinOrWaiterAnyValue WinOrWaiterStrategy = iota
	// WinOrWaiterMustBeLater forces a calculation after the last arriving thread.
	// WinOrWaiter caclulations are serialized, ie. a new calculation does not start prior to
	// the conlusion of the previous calulation
	WinOrWaiterMustBeLater
)

// WinOrWaiter picks a winner thread to carry out some task used by many threads.
//   - threads in WinOrWait for an idle WinorWaiter may become winners completing the task
//   - threads in WinOrWait while a calculation is in progress are held waiting using
//     RWLock and atomics until the calculation completes
//   - the calculation is completed on demand, but only by the first requesting thread
type WinOrWaiterCore struct {
	// calculator if the function making the calculation
	calculator func() (err error)
	// calculation strategy for this WinOrWaiter
	//	- WinOrWaiterAnyValue WinOrWaiterMustBeLater
	strategy WinOrWaiterStrategy
	// context used for cancellation, may be nil
	ctx context.Context

	// isCalculationPut indicates that calculation field has value. atomic access
	isCalculationPut AtomicBool
	// calculationPut makes threads wait until calculation has value
	calculationPut Once
	// calculation allow to wait for the result of a winner calculation
	//	- winner holds lock.Lock until the calculation is complete
	//	- loser threads wait for lock.RLock to check the result
	calculation AtomicReference[AwaitableCalculation[time.Time]]

	// winnerPicker picks winner thread using atomic access
	//	- winner is the thread that on Set gets wasNotSet true
	//	- true while a winner calculates next data value
	//	- set to zero when winnerFunc returns
	winnerPicker AtomicBool
}

// WinOrWaiter returns a semaphore used for completing an on-demand task by
// the first thread requesting it, and that result shared by subsequent threads held
// waiting for the result.
//   - strategy: WinOrWaiterAnyValue WinOrWaiterMustBeLater
//   - ctx allows foir cancelation of the WinOrWaiter
func NewWinOrWaiterCore(strategy WinOrWaiterStrategy, calculator func() (err error), ctx ...context.Context) (winOrWaiter *WinOrWaiterCore) {
	if !strategy.IsValid() {
		panic(perrors.ErrorfPF("Bad WinOrWaiter strategy: %s", strategy))
	}
	if calculator == nil {
		panic(perrors.ErrorfPF("calculator function cannot be nil"))
	}
	var ctx0 context.Context
	if len(ctx) > 0 {
		ctx0 = ctx[0]
	}
	return &WinOrWaiterCore{
		strategy:   strategy,
		calculator: calculator,
		ctx:        ctx0,
	}
}

// WinOrWaiter picks a winner thread to carry out some task used by many threads.
//   - threads arriving to an idle WinorWaiter are winners that complete the task
//   - threads arriving to a WinOrWait in progress are held waiting at RWMutex
//   - the task is completed on demand, but only by the first thread requesting it
func (ww *WinOrWaiterCore) WinOrWait() (err error) {

	// arrivalTime is the time this thread arrived
	var arrivalTime = time.Now()

	// ensure WinOrWaiter has calculator function
	if ww == nil || ww.calculator == nil {
		err = perrors.NewPF("WinOrWait for nil or uninitialized WinOrWaiter")
		return
	}
	// seenCalculation is the calculation present when this thread arrived.
	// seenCalculation may be nil
	var seenCalculation = ww.calculation.Get()

	// ensure that ww.calculation holds a calculation
	if ww.isCalculationPut.IsFalse() {

		// invocation prior to first calculation started
		// start the first calculation, or wait for it to be started if another thread already started it
		var didOnce bool
		if didOnce, _, err = ww.calculationPut.DoErr(ww.winnerFunc); didOnce {
			DidOnceReturn.Inc()
			return // thread did initial winner calculation return
		}
		err = nil // subsequent threads do not report possible error
	}

	// wait for late-enough data
	var calculation *AwaitableCalculation[time.Time]
	for {

		// check for valid calculation result
		calculation = ww.calculation.Get()
		// calculation.Result may block
		if result, isValid := calculation.Result(); isValid {
			switch ww.strategy {
			case WinOrWaiterAnyValue:
				if calculation != seenCalculation {
					NewCalculationReturn.Inc()
					return // any new valid value accepted return
				}
			case WinOrWaiterMustBeLater:
				if !result.Before(arrivalTime) {
					AfterReturn.Inc()
					Arrival.SetTime(arrivalTime)
					Result.SetTime(result)
					// arrival time the same or after dataVersionNow
					return // must be later and the data version is of a later time than when this thread arrived return
				}
			}
		}

		// ensure data processing is in progress
		if isWinner := ww.winnerPicker.Set(); isWinner {
			WinnerReturn.Inc()
			return ww.winnerFunc() // this thread completed the task return
		}

		// check context cancelation
		if ww.IsCancel() {
			CancelReturn.Inc()
			return // context canceled return
		}
	}
}

func (ww *WinOrWaiterCore) IsCancel() (isCancel bool) {
	return ww.ctx != nil && ww.ctx.Err() != nil
}

func (ww *WinOrWaiterCore) winnerFunc() (err error) {
	ww.winnerPicker.Set()
	defer ww.winnerPicker.Clear()

	// get calculation
	var calculation = NewAwaitableCalculation[time.Time]()
	ww.calculation.Put(calculation)
	ww.isCalculationPut.Set()

	// calculate
	result := time.Now()
	defer calculation.End(result, &err)
	Calc.Inc()
	_, err = RecoverInvocationPanicErr(ww.calculator)

	return
}

type WinOrWaiterStrategy uint8

func (ws WinOrWaiterStrategy) String() (s string) {
	return winOrWaiterSet.StringT(ws)
}

func (ws WinOrWaiterStrategy) IsValid() (isValid bool) {
	return winOrWaiterSet.IsValid(ws)
}

var winOrWaiterSet = sets.NewSet(sets.NewElements[WinOrWaiterStrategy](
	[]sets.SetElement[WinOrWaiterStrategy]{
		{ValueV: WinOrWaiterAnyValue, Name: "anyValue"},
		{ValueV: WinOrWaiterMustBeLater, Name: "mustBeLater"},
	}))
