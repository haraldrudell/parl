/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// serialdo invokes method in sequence
type SerialDo struct {
	inChan     chan *time.Time
	inChanLock sync.Mutex
	isBusy     AtomicBool
	isShutdown AtomicBool
	thunk      func()
	cbFunc     SerialDoFunc
	ErrCh      chan error
	ID         string
	Wg         sync.WaitGroup
	ctx        context.Context
}

// NewSerialDo SerialDo. errors on sdo.ErrCh
func NewSerialDo(thunk func(), eventReceiver SerialDoFunc, ctx context.Context) (sdo *SerialDo) {
	if thunk == nil {
		panic(Errorf("NewSerialDo with thunk nil"))
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if eventReceiver == nil {
		eventReceiver = func(e SerialDoEvent, s *SerialDo, t *time.Time) {}
	}
	ID := strconv.FormatUint(atomic.AddUint64(&serialDoNo, 1), 10)
	sdo = &SerialDo{
		inChan: make(chan *time.Time, 1), inChanLock: sync.Mutex{},
		thunk: thunk, cbFunc: eventReceiver, ErrCh: make(chan error),
		ID: ID, ctx: ctx,
	}
	sdo.Wg.Add(1)
	go sdo.inReader()
	return
}

type SerialDoEvent uint8

type SerialDoFunc func(SerialDoEvent, *SerialDo, *time.Time)

var serialDoNo uint64

const (
	SerialDoReady         = 0 + iota
	SerialDoLaunch        // from idle, now time
	SerialDoPending       // queued up invocation, request time
	SerialDoPendingLaunch // launch of pending invocation, request time
	SerialDoIdle          // busy since
)

// Invoke thunk serially, maximum queue one invocation, drop additional invocation requests prior to idle. non-blocking Thread-safe
func (sdo *SerialDo) Do(now time.Time) (nowPending bool) {
	if sdo.isShutdown.IsTrue() {
		panic(Errorf("SerialDo#%s: Do after shutdown", sdo.ID))
	}
	if len(sdo.inChan) > 0 || !sdo.sendInChan(&now) {
		return
	}
	nowPending = sdo.isBusy.IsTrue()
	if nowPending {
		sdo.cbFunc(SerialDoPending, sdo, &now)
	}
	return
}

func (sdo *SerialDo) sendInChan(now *time.Time) bool {
	sdo.inChanLock.Lock()
	defer sdo.inChanLock.Unlock()
	if len(sdo.inChan) > 0 || sdo.isShutdown.IsTrue() {
		return false
	}
	sdo.inChan <- now
	return true
}

func (sdo *SerialDo) Shutdown() {
	if !sdo.isShutdown.Set() {
		return // was already set
	}
	sdo.inChanLock.Lock()
	defer sdo.inChanLock.Unlock()
	close(sdo.inChan) // closes run thread
}

func (sdo *SerialDo) inReader() {
	defer sdo.Wg.Done()
	defer close(sdo.ErrCh)
	defer Recover("SerialDo#"+sdo.ID+".inReader", nil, func(e error) { sdo.ErrCh <- e })

	sdo.cbFunc(SerialDoReady, sdo, nil)
	var launchTime *time.Time
	var ok bool
	for {
		select {
		case launchTime, ok = <-sdo.inChan:
		case _, ok = <-sdo.ctx.Done():
			sdo.Shutdown()
		}
		if !ok {
			break
		}

		// data from inChan: launch
		sdo.cbFunc(SerialDoLaunch, sdo, launchTime)
		for {
			sdo.isBusy.Set()
			ch := make(chan struct{})
			sdo.Wg.Add(1)
			go func() {
				defer sdo.Wg.Done()
				defer close(ch)
				sdo.thunk()
			}()
			select {
			case <-ch:
			case _, ok = <-sdo.ctx.Done():
				sdo.Shutdown()
			}
			if !ok || // application is terminating
				len(sdo.inChan) < 1 { // no pending requests: enter idle state
				sdo.isBusy.Clear()
				break
			}
			pending := <-sdo.inChan
			sdo.cbFunc(SerialDoPendingLaunch, sdo, pending)
		}
		if !ok {
			break
		}
		sdo.cbFunc(SerialDoIdle, sdo, launchTime) // now idle, busy since launchTime
	}
}
