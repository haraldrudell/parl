/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"sync"
	"time"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/ptime"
	"github.com/haraldrudell/parl/sets"
)

const (
	// WinOrWaiterAnyValue causes a thread to accept any calculated value
	WinOrWaiterAnyValue WinOrWaiterStrategy = iota + 1
	// WinOrWaiterMustBeLater forces a calculation after the last arriving thread.
	// WinOrWaiter caclulations are serialized, ie. a new calculation does not start prior to
	// the conlusion of the previous calulation
	WinOrWaiterMustBeLater
)

// WinOrWaiter picks a winner thread to carry out some task used by many threads.
//   - threads arriving to an idle WinorWaiter are winners that complete the task
//   - After a winning thread completes the task, it invokes WinnerDone
//   - threads arriving to a WinOrWait in progress are held waiting until WinnerDone
//   - the task is completed on demand, but only by the first thread requesting it
type WinOrWaiterCore struct {
	cond sync.Cond
	// calculation strategy for this WinOrWaiter
	//	- WinOrWaiterAnyValue WinOrWaiterMustBeLater
	strategy WinOrWaiterStrategy
	// context used for cancellation, may be nil
	ctx context.Context

	// dataVersion indicates the available data version with atomic access
	//	- data version is the time scanning of the current valid version of data began
	//	- zero value if no data version has been completed
	//	- only updated by the winner on its completion using winnerFunc
	dataVersion ptime.EpochValue
	// winnerPicker picks winner thread using atomic access
	//	- winner is the thread that on Set gets wasNotSet true
	//	- true while a winner calculates next data value
	//	- set to zero after winnerFunc invoked
	winnerPicker AtomicBool
	// calculationStart is the time of starting the last initiated calculation with atomic access
	calculationStart ptime.EpochValue
}

// WinOrWaiter returns a semaphore used for completing an on-demand task by
// the first thread requesting it, and that result shared by subsequent threads held
// waiting for the result.
//   - strategy: WinOrWaiterAnyValue WinOrWaiterMustBeLater
//   - ctx allows foir cancelation of the WinOrWaiter
func NewWinOrWaiterCore(strategy WinOrWaiterStrategy, ctx ...context.Context) (winOrWaiter *WinOrWaiterCore) {
	var ctx0 context.Context
	if len(ctx) > 0 {
		ctx0 = ctx[0]
	}
	if !strategy.IsValid() {
		panic(perrors.ErrorfPF("Bad WInOrWaiter strategy: %s", strategy))
	}
	return &WinOrWaiterCore{
		cond:     *sync.NewCond(&sync.Mutex{}),
		strategy: strategy,
		ctx:      ctx0,
	}
}

// WinOrWaiter picks a winner thread to carry out some task used by many threads.
//   - threads arriving to an idle WinorWaiter are winners that complete the task
//   - After a winning thread completes the task, it invokes WinnerDone
//   - threads arriving to a WinOrWait in progress are held waiting until WinnerDone
//   - the task is completed on demand, but only by the first thread requesting it
func (ww *WinOrWaiterCore) WinOrWait() (winnerFunc func(errp *error)) {
	checkWinOrWaiter(ww)
	ww.cond.L.Lock()
	defer ww.cond.L.Unlock()

	// the time this thread arrived
	var arrivalTime = time.Now()
	// the data version available when this thread arrived
	var lastSeenDataVersion = ww.dataVersion.Get().Time()

	// wait for a data update
	for {

		// check context
		if ww.IsCancel() {
			return // context canceled return
		}

		// if there has been a data update since this thread arrived
		dataVersionNow := ww.dataVersion.Get().Time()
		// if we have data and it has changed since arrival…
		if !dataVersionNow.IsZero() && !lastSeenDataVersion.Equal(dataVersionNow) {
			switch ww.strategy {
			case WinOrWaiterAnyValue:
				return // any new valid value accepted return
			case WinOrWaiterMustBeLater:
				if !arrivalTime.Before(dataVersionNow) {
					// arrival time the same or after dataVersionNow
					return // must be later and the data version is of a later time than when this thread arrived return
				}
			}
		}
		lastSeenDataVersion = dataVersionNow // absorb any changes

		// ensure data processing is in progress
		if isWinner := ww.winnerPicker.Set(); isWinner {

			// this thread is a winner!
			ww.calculationStart.SetTime()
			winnerFunc = ww.winnerFunc
			return // this thread is a winner: do task return
		}

		// wait for any updates
		ww.cond.Wait()
	}
}

// Invalidate invalidates any completed calculation.
// A calculation in progress may still be accepted.
func (ww *WinOrWaiterCore) Invalidate() {
	checkWinOrWaiter(ww)
	// invalidate current data version
	// for performance, important to do outside of lock
	ww.dataVersion.Set(0)

	ww.cond.L.Lock()
	defer ww.cond.L.Unlock()

	ww.cond.Broadcast()
}

func (ww *WinOrWaiterCore) IsCancel() (isCancel bool) {
	checkWinOrWaiter(ww)
	return ww.ctx != nil && ww.ctx.Err() != nil
}

func (ww *WinOrWaiterCore) winnerFunc(errp *error) {

	// if successful, update data version
	// for performance, important to do outside of lock
	// when dataVersion is updated, waiting threads will begin to return
	if errp == nil || *errp == nil {
		ww.dataVersion.Set(ww.calculationStart.Get())
	}

	ww.cond.L.Lock()
	defer ww.cond.L.Unlock()

	// broadcast to wake all waiting threads
	ww.cond.Broadcast()

	// allow for next winner to be picked
	ww.winnerPicker.Clear()
}

func checkWinOrWaiter(ww *WinOrWaiterCore) {
	if ww == nil {
		panic(perrors.NewPF("use of nil WinOrWaiterCore"))
	} else if ww.cond.L == nil {
		panic(perrors.NewPF("use of uninitialized WinOrWaiterCore"))
	}
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
