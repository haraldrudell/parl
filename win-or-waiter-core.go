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
	// available data version: the time scanning of the current version of data began.
	// zero-time if no data version has been completed
	tVersion time.Time
	// the time of starting the last initiated calculation
	tStart time.Time
	// calculation strategy for this WinOrWaiter
	//	- WinOrWaiterAnyValue WinOrWaiterMustBeLater
	strategy WinOrWaiterStrategy
	// context used for cancellation, may be nil
	ctx context.Context
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
func (ww *WinOrWaiterCore) WinOrWait() (isWinner bool) {
	checkWinOrWaiter(ww)
	ww.cond.L.Lock()
	defer ww.cond.L.Unlock()

	// the time this thread arrived
	var tThis = time.Now()
	// the data version available when this thread arrived
	var tVersion = ww.tVersion

	// wait for a data update
	for {

		// check context
		if ww.IsCancel() {
			return // context canceled return
		}

		// if there has been a data update since this thread arrived
		if !tVersion.Equal(ww.tVersion) {
			switch ww.strategy {
			case WinOrWaiterAnyValue:
				if !ww.tVersion.IsZero() {
					return // any value accepted and a value is valid return
				}
			case WinOrWaiterMustBeLater:
				if !tThis.Before(ww.tVersion) {
					return // must be later and the data version is of a later time than when this thread arrived return
				}
			}
		}
		tVersion = ww.tVersion // absorb any changes

		// ensure data processing is in progress
		if isWinner = tVersion.Equal(ww.tStart); isWinner {
			ww.tStart = time.Now()
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
	ww.cond.L.Lock()
	defer ww.cond.Broadcast()
	defer ww.cond.L.Unlock()

	if ww.tVersion.Equal(ww.tStart) {
		ww.tStart = time.Time{} // no calculation in progress, ensure ww.tStart and ww.tVerison the same
	}
	// indicate that data value is not valid
	ww.tVersion = time.Time{}
}

// WinOrWait allows a winning thread to announce completion of the task.
// Deferrable.
func (ww *WinOrWaiterCore) WinnerDone(errp *error) {
	checkWinOrWaiter(ww)
	ww.cond.L.Lock()
	defer ww.cond.Broadcast()
	defer ww.cond.L.Unlock()

	isError := errp != nil && *errp != nil

	// if no error, record the new data version
	if !isError {
		ww.tVersion = ww.tStart
	} else {

		// on error, indicate that data scanning is no longer in progress
		ww.tStart = ww.tVersion
	}
}

func (ww *WinOrWaiterCore) IsCancel() (isCancel bool) {
	checkWinOrWaiter(ww)
	return ww.ctx != nil && ww.ctx.Err() != nil
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

var winOrWaiterSet = sets.NewSet(sets.NewElements[WinOrWaiterStrategy](
	[]sets.SetElement[WinOrWaiterStrategy]{
		{ValueV: WinOrWaiterAnyValue, Name: "anyValue"},
		{ValueV: WinOrWaiterMustBeLater, Name: "mustBeLater"},
	}))
