/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	shortTime = time.Millisecond
)

func TestAccessManagerAtomic(t *testing.T) {
	var exp1 = ticketDelta
	var exp2 = uint64(0)

	var manager accessManager
	var actual uint64

	// start atomic operation
	var isLockAccess = manager.RequestAccess()
	if isLockAccess {
		t.Error("isLockAccess true")
	}
	actual = atomic.LoadUint64(&manager.before)
	if actual != exp1 {
		t.Errorf("before1: %d exp %d", actual, exp1)
	}

	// end atomic operation
	manager.RelinquishAccess(isLockAccess)
	actual = atomic.LoadUint64(&manager.before)
	if actual != exp2 {
		t.Errorf("before2: %d exp %d", actual, exp2)
	}
}

type UseLockTester struct {
	lockIsReady, waitToReleaseLock, lockDone sync.WaitGroup
	hasLock                                  chan struct{}
	requestIsReady, requestDone              sync.WaitGroup
	hasTicket                                chan struct{}
}

func NewUseLockTester() (lockTester *UseLockTester) { return &UseLockTester{} }
func (u *UseLockTester) Lock(manager *accessManager) {
	u.lockIsReady.Add(1)
	u.waitToReleaseLock.Add(1)
	u.lockDone.Add(1)
	u.hasLock = make(chan struct{})
	go u.lock(manager)
}
func (u *UseLockTester) lock(manager *accessManager) {
	defer u.lockDone.Done()

	u.lockIsReady.Done()
	manager.Lock()
	close(u.hasLock)
	u.waitToReleaseLock.Wait()
	manager.Unlock()
}
func (u *UseLockTester) Request(manager *accessManager) {
	u.requestIsReady.Add(1)
	u.requestDone.Add(1)
	u.hasTicket = make(chan struct{})
	go u.request(manager)
}
func (u *UseLockTester) request(manager *accessManager) {
	defer u.requestDone.Done()

	u.requestIsReady.Done()
	var ticket = manager.RequestAccess()
	close(u.hasTicket)
	manager.RelinquishAccess(ticket)
}

func TestAccessManagerLock(t *testing.T) {

	var manager accessManager
	var lockTester = NewUseLockTester()
	var timer *time.Timer

	// have ongoing atomic access: blocks Lock
	var ticket = manager.RequestAccess()

	// have thread wait for lock access
	lockTester.Lock(&manager)
	lockTester.lockIsReady.Wait()
	select {
	case <-lockTester.hasLock:
		t.Error("hasLock prematurely")
	default:
	}

	// end atomic access: enter lock access
	manager.RelinquishAccess(ticket)
	timer = time.NewTimer(shortTime)
	select {
	case <-timer.C:
		t.Error("hasLock not enabled")
	case <-lockTester.hasLock:
		timer.Stop()
	}

	// attempt lock access: should be blocked
	lockTester.Request(&manager)
	lockTester.requestIsReady.Wait()
	select {
	case <-lockTester.hasTicket:
		t.Error("hasTicket prematurely")
	default:
	}

	// make lock available: lock access should happen
	lockTester.waitToReleaseLock.Done()
	timer = time.NewTimer(shortTime)
	select {
	case <-timer.C:
		t.Error("hasTicket not enabled")
	case <-lockTester.hasTicket:
		timer.Stop()
	}

	lockTester.lockDone.Wait()
	lockTester.requestDone.Wait()
}
