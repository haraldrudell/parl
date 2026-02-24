/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl_test

import (
	"fmt"
	"testing"

	"github.com/haraldrudell/parl"
)

func TestSpinLock_Lock(t *testing.T) {
	var (
		lock    parl.SpinLock
		thread1 = newProgressThread(&lock)
		thread2 = newProgressThread(&lock)
		state   threadState
	)

	// start threads
	go thread1.createThread()
	<-thread1.awaitLockHeldCh
	go thread2.createThread()
	<-thread2.runningCh
	// thread1 holds the lock, thread2 is trying to acquire lock

	// thread1 state should be holdingAtRelease
	state = thread1.state.Load()
	if state != holdingAtRelease {
		t.Fatalf("FAIL bad state thread1 holding lock: %s exp %s", state, holdingAtRelease)
	}

	// thread2 state should be holdingAtLock
	state = thread2.state.Load()
	if state != holdingAtLock {
		t.Fatalf("FAIL bad state thread2 acquiring lock: %s exp %s", state, holdingAtRelease)
	}

	// release thread1
	close(thread1.releaseCh)
	// await thread1 exit
	<-thread1.exitCh

	// await thread2 acquiring the lock
	<-thread2.awaitLockHeldCh

	// thread2 state should be holdingAtRelease
	state = thread2.state.Load()
	if state != holdingAtRelease {
		t.Fatalf("FAIL bad state thread2 holding lock: %s exp %s", state, holdingAtRelease)
	}

	// release thread2
	close(thread2.releaseCh)
	// await thread2 exit
	<-thread2.exitCh
}

type progressThread struct {
	// lock holds reference to shared lock
	lock *parl.SpinLock
	// runningCh closes once the thread is running
	runningCh chan struct{}
	// exitCh closes upon thread exit
	exitCh chan struct{}
	// awaitLockHeldCh closes upon lock acquired
	awaitLockHeldCh chan struct{}
	// closing releaseCh instructs the thread to release the lock
	releaseCh chan struct{}
	// state is [uninitialized] [holdingAtLock] [holdingAtRelease] [exited]
	state parl.Atomic32[threadState]
}

func newProgressThread(lock *parl.SpinLock) (t *progressThread) {
	return &progressThread{
		lock:            lock,
		runningCh:       make(chan struct{}),
		exitCh:          make(chan struct{}),
		awaitLockHeldCh: make(chan struct{}),
		releaseCh:       make(chan struct{}),
	}
}

func (t *progressThread) createThread() {
	defer close(t.exitCh)
	defer t.state.Store(exited)

	t.state.Store(holdingAtLock)
	close(t.runningCh)
	t.lock.Lock()
	t.state.Store(holdingAtRelease)
	close(t.awaitLockHeldCh)
	<-t.releaseCh
	t.lock.Unlock()
}

const (
	// the thread has not been created or started running
	uninitialized threadState = iota
	// the thread will attempt to acquire the lock
	holdingAtLock
	// the thread will release the lock
	holdingAtRelease
	// the thread has exited
	exited
)

// status value for where the therad is
//   - [uninitialized][holdingAtLock] [holdingAtRelease] [exited]
type threadState uint8

func (s threadState) String() (s2 string) {
	switch s {
	case uninitialized:
		return "uninitialized"
	case holdingAtLock:
		return "holdingAtLock"
	case holdingAtRelease:
		return "holdingAtRelease"
	case exited:
		return "exited"
	default:
		return fmt.Sprintf("?threadState:%d", s)
	}

}
