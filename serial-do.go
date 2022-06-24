/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/haraldrudell/parl/perrors"
)

// SerialDo serializes a thunk.
//
type SerialDo struct {
	thunk         func(at time.Time)
	eventReceiver func(event *SerialDoEvent)
	errFn         func(err error)

	lock         sync.Mutex
	busySince    time.Time // behind lock
	pendingSince time.Time // behind lock

	sdoID SerialDoID
	wg    WaitGroup
	ctx   context.Context
}

// NewSerialDo serializes .Do invocations
// eventFn must be thread-safe.
// errFn must be thread-safe.
func NewSerialDo(
	thunk func(at time.Time),
	eventFn func(event *SerialDoEvent),
	errFn func(err error),
	ctx context.Context) (sdo *SerialDo) {

	if thunk == nil {
		panic(perrors.Errorf("NewSerialDo with thunk nil"))
	}
	if errFn == nil {
		panic(perrors.Errorf("NewSerialDo with errFn nil"))
	}
	return &SerialDo{
		thunk:         thunk,
		eventReceiver: eventFn,
		errFn:         errFn,
		sdoID:         serialDoID.ID(),
		ctx:           ctx,
	}
}

// Do invokes the thunk serially
// if SerialDo is idle, Do launches thunk via a thread
// if SerialDo is busy, Do makes it pending
// if SerialDo is already pending, Do does nothing
// Do returns true if SerialDo state is pending
// Do is non-blocking and thread-safe
func (sdo *SerialDo) Do(now ...time.Time) (isPending bool) {

	// check state
	if sdo.ctx.Err() != nil {
		panic(perrors.Errorf("SerialDo#%s: Do after cancel", sdo.sdoID))
	}

	// ensure now0 non-zero time
	var now0 time.Time
	if len(now) > 0 {
		now0 = now[0]
	}
	if now0.IsZero() {
		now0 = time.Now()
	}

	var typ SerialDoType
	if typ, isPending = sdo.performDo(now0); !typ.IsValid() {
		return // was already pending: ignore Do invocation
	}

	// SerialDoPending: pending at now0
	// SerialDoLaunch: busy at now0
	// send event
	sdo.invokeEventFn(typ, now0)

	return // state changed return
}

func (sdo *SerialDo) State() (
	busySince time.Time,
	pendingSince time.Time,
	isCancel bool,
	isWaitComplete bool,
) {
	if isCancel = sdo.ctx.Err() != nil; isCancel {
		isWaitComplete = sdo.wg.IsZero()
	}

	sdo.lock.Lock()
	defer sdo.lock.Unlock()

	busySince = sdo.busySince
	pendingSince = sdo.pendingSince
	return
}

func (sdo *SerialDo) Wait() {

	// wait for shutdown
	<-sdo.ctx.Done()

	// wait for thread to exit
	sdo.wg.Wait()
}

func (sdo *SerialDo) String() (s string) {
	s = "sdo#" + sdo.sdoID.String()
	busySince, pendingSince, isCancel, isWaitComplete := sdo.State()

	// check if done
	if isWaitComplete {
		return s + " done"
	}

	return fmt.Sprintf("%s busy: %s pend: %s can: %t",
		s, Short(busySince), Short(pendingSince),
		isCancel,
	)
}

func (sdo *SerialDo) performDo(now time.Time) (typ SerialDoType, isPending bool) {
	sdo.lock.Lock()
	defer sdo.lock.Unlock()

	// if already pending, ignore
	if !sdo.pendingSince.IsZero() {
		isPending = true
		return // already pending
	}

	// if busy, mark pending
	if !sdo.busySince.IsZero() {
		sdo.pendingSince = now
		return SerialDoPending, true
	}

	// launch thread
	sdo.busySince = now
	sdo.wg.Add(1)
	go sdo.doThread(now)
	typ = SerialDoLaunch
	return
}

func (sdo *SerialDo) invokeEventFn(typ SerialDoType, t time.Time) {
	defer Recover(Annotation(), nil, sdo.errFn)

	if sdo.eventReceiver == nil {
		return
	}

	sdo.eventReceiver(NewSerialDoEvent(typ, t, sdo))
}

func (sdo *SerialDo) invokeThunk(at time.Time) {
	defer Recover(Annotation(), nil, sdo.errFn)

	sdo.thunk(at)
}

func (sdo *SerialDo) doThread(at time.Time) {
	defer sdo.wg.Done()
	defer Recover(Annotation(), nil, sdo.errFn)

	for {
		sdo.invokeThunk(at)

		var typ SerialDoType
		typ, at = sdo.checkForMoreDo()
		sdo.invokeEventFn(typ, at)
		if typ == SerialDoIdle {
			return
		}
	}
}

var timeTimeZeroValue time.Time

func (sdo *SerialDo) checkForMoreDo() (typ SerialDoType, t time.Time) {
	sdo.lock.Lock()
	defer sdo.lock.Unlock()

	// check for cancel
	if sdo.ctx.Err() != nil {
		t = sdo.busySince
		sdo.busySince = timeTimeZeroValue
		sdo.pendingSince = timeTimeZeroValue
		typ = SerialDoIdle
		return
	}

	// if pending, relaunch thunk
	if !sdo.pendingSince.IsZero() {
		t = sdo.pendingSince
		sdo.pendingSince = timeTimeZeroValue
		typ = SerialDoPendingLaunch
		return
	}

	// enter idle state
	t = sdo.busySince
	sdo.busySince = timeTimeZeroValue
	typ = SerialDoIdle
	return
}
