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

type SerialDoCore struct {
	thunk func(at time.Time)

	cond         *sync.Cond
	pendingSince time.Time // behind lock
	busySince    time.Time // behind lock

	errFn func(err error)
	ctx   context.Context
}

func NewSerialDoCore(thunk func(at time.Time), errFn func(err error), ctx context.Context) (serialDoCore *SerialDoCore) {
	return &SerialDoCore{
		thunk: thunk,
		cond:  sync.NewCond(&sync.Mutex{}),
		errFn: errFn,
		ctx:   ctx,
	}
}

func (sdo *SerialDoCore) Do(now ...time.Time) (isPending bool, isShutdown bool) {

	// check state
	if sdo.ctx.Err() != nil {
		isShutdown = true
		return // isShutdown
	}

	// obtain request time in now0
	var now0 time.Time
	if len(now) > 0 {
		now0 = now[0]
	}
	if now0.IsZero() {
		now0 = time.Now()
	}

	sdo.cond.L.Lock()
	defer sdo.cond.L.Unlock()

	// if already pending, ignore
	if !sdo.pendingSince.IsZero() {
		isPending = true
		return // no-op: already pending
	}

	// if busy, mark pending
	if !sdo.busySince.IsZero() {
		sdo.pendingSince = now0
		isPending = true
		return
	}

	// launch thread
	sdo.busySince = now0
	go sdo.doThread(now0)

	return
}

func (sdo *SerialDoCore) Wait(at time.Time) {
	<-sdo.ctx.Done()
	sdo.cond.L.Lock()
	defer sdo.cond.L.Unlock()

	for {
		if sdo.busySince.IsZero() {
			return // cancel and no thread: done
		}

		sdo.cond.Wait()
	}
}

func (sdo *SerialDoCore) doThread(at time.Time) {
	defer Recover(Annotation(), nil, sdo.errFn)

	for {
		sdo.invokeThunk(at)

		if at = sdo.checkForMoreDo(); at.IsZero() {
			return
		}
	}
}

func (sdo *SerialDoCore) checkForMoreDo() (at time.Time) {
	sdo.cond.L.Lock()
	defer sdo.cond.L.Unlock()

	// check for cancel
	if sdo.ctx.Err() != nil {
		sdo.busySince = time.Time{}
		sdo.pendingSince = time.Time{}
		return
	}

	// if pending, relaunch thunk
	if !sdo.pendingSince.IsZero() {
		at = sdo.pendingSince
		sdo.pendingSince = time.Time{}
		return
	}

	// idle: thread should exit
	sdo.busySince = time.Time{}
	sdo.cond.Signal()
	return
}

func (sdo *SerialDoCore) invokeThunk(at time.Time) {
	defer Recover(Annotation(), nil, sdo.errFn)

	sdo.thunk(at)
}
