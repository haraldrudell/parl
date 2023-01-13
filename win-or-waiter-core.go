/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"sync"
	"time"
)

// WinOrWaiter picks a winner thread to carry out some task used by many threads.
//   - threads arriving to an idle WinorWaiter are winners that complete the task
//   - After a winning thread completes the task, it invokes WinnerDone
//   - threads arriving to a WinOrWait in progress are held waiting until WinnerDone
//   - the task is completed on demand, but only by the first thread requesting it
type WinOrWaiterCore struct {
	sync.Cond
	// data version: the time scanning of the current version of data began.
	// zero-time if no data version has been completed
	tVersion time.Time
	// the time of starting the last initiated scan
	tStart      time.Time
	mustBeLater bool
	ctx         context.Context
}

// WinOrWaiter returns a semaphore used for completing an on-demand task by
// the first thread requesting it, and that result shared by subsequent threads held
// waiting for the result.
func NewWinOrWaiterCore(mustBeLater bool, ctx ...context.Context) (winOrWaiter *WinOrWaiterCore) {
	var ctx0 context.Context
	if len(ctx) > 0 {
		ctx0 = ctx[0]
	}
	return &WinOrWaiterCore{
		Cond:        *sync.NewCond(&sync.Mutex{}),
		mustBeLater: mustBeLater,
		ctx:         ctx0,
	}
}

// WinOrWaiter picks a winner thread to carry out some task used by many threads.
//   - threads arriving to an idle WinorWaiter are winners that complete the task
//   - After a winning thread completes the task, it invokes WinnerDone
//   - threads arriving to a WinOrWait in progress are held waiting until WinnerDone
//   - the task is completed on demand, but only by the first thread requesting it
func (ww *WinOrWaiterCore) WinOrWait() (isWinner bool) {
	ww.Cond.L.Lock()
	defer ww.Cond.L.Unlock()

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
			if !ww.mustBeLater {
				return // data change occurred and time does not matter return
			}

			if !tThis.Before(ww.tVersion) {
				return // the data version is of a later time than when this thread arrived return
			}
		}

		// ensure data processing is in progress
		if isWinner = tVersion.Equal(ww.tStart); isWinner {
			ww.tStart = time.Now()
			return // this thread is a winner: do task return
		}

		// wait for any updates
		ww.Cond.Wait()
	}
}

// WinOrWait allows a winning thread to announce completion of the task.
// Deferrable.
func (ww *WinOrWaiterCore) WinnerDone(errp *error) {
	ww.Cond.L.Lock()
	defer ww.Cond.Broadcast()
	defer ww.Cond.L.Unlock()

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
	return ww.ctx != nil && ww.ctx.Err() != nil
}
