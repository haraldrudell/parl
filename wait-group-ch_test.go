/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWaitGroupCh(t *testing.T) {
	const (
		adds0            = 2
		adds1            = 1
		currentCountExp0 = 1
		sExp             = "waitGroupCh_count:0(adds:0)_isWaiting:false_isClosed:false"
	)

	var (
		awaitableCh              AwaitableCh
		isClosed, isExit, isZero bool
		currentCount, totalAdds  int
		s                        string
	)

	// Add() Done() DoneBool() Wait()
	// Ch() Count() Counts() IsZero()
	// Reset() String()
	var w WaitGroupCh
	var reset = func() {
		w = WaitGroupCh{}
	}

	// Ch should return a non-closed channel
	reset()
	w.Add(adds1)
	awaitableCh = w.Ch()
	select {
	case <-awaitableCh:
		isClosed = true
	default:
		isClosed = false
	}
	if isClosed {
		t.Error("Ch is closed")
	}

	// Add-Done should close channel
	reset()
	w.Add(adds1)
	awaitableCh = w.Ch()
	w.Done()
	select {
	case <-awaitableCh:
		isClosed = true
	default:
		isClosed = false
	}
	if !isClosed {
		t.Error("Ch not closed")
	}

	// Count should return zeroes
	reset()
	currentCount, totalAdds = w.Counts()
	if currentCount != 0 {
		t.Errorf("Count currentCount %d exp 0", currentCount)
	}
	if totalAdds != 0 {
		t.Errorf("Count totalAdds %d exp 0", totalAdds)
	}

	// Count should reflect Adds Dones
	//	- Add() Done()
	reset()
	w.Add(adds0)
	w.Done()
	currentCount, totalAdds = w.Counts()
	if currentCount != currentCountExp0 {
		t.Errorf("Count currentCount %d exp %d", currentCount, currentCountExp0)
	}
	if totalAdds != adds0 {
		t.Errorf("Count totalAdds %d exp %d", totalAdds, adds0)
	}

	// DoneBool
	reset()
	w.Add(adds0)
	isExit = w.DoneBool()
	if isExit {
		t.Error("DoneBool isExit true")
	}
	isExit = w.DoneBool()
	if !isExit {
		t.Error("DoneBool isExit false")
	}

	// Reset should reset adds
	reset()
	w.Add(adds0)
	currentCount, totalAdds = w.Counts()
	_ = currentCount
	if totalAdds == 0 {
		t.Error("totalAdds zero")
	}
	w.Reset()
	currentCount, totalAdds = w.Counts()
	_ = currentCount
	if totalAdds != 0 {
		t.Error("totalAdds not zero")
	}

	// IsZero should be false after Add
	reset()
	w.Add(adds1)
	isZero = w.IsZero()
	if isZero {
		t.Error("IsZero true")
	}
	// IsZero should be true after Add-Done
	w.Done()
	isZero = w.IsZero()
	if !isZero {
		t.Error("IsZero false")
	}

	// String()
	reset()
	s = w.String()
	if s != sExp {
		t.Errorf("String %q exp %q", s, sExp)
	}
}

func TestWaitGroupChWait(t *testing.T) {
	var adds1 = 1
	var shortTime = time.Millisecond

	var isReady sync.WaitGroup
	var isDone chan struct{}
	var isWaitReturn atomic.Bool
	var timer *time.Timer

	// Add() Ch() Count() Done() DoneBool() IsZero() Reset() String()
	// Wait()
	var w WaitGroupCh
	var reset = func() {
		w = WaitGroupCh{}
	}

	// Wait waits until counter zero
	reset()
	w.Add(adds1)
	isWaitReturn.Store(false)
	isReady = sync.WaitGroup{}
	isReady.Add(1)
	isDone = make(chan struct{})
	go waiter(&w, &isWaitReturn, &isReady, isDone)
	isReady.Wait()
	<-time.NewTimer(shortTime).C
	if isWaitReturn.Load() {
		t.Error("Wait returned prematurely")
	}
	w.Done()
	// race condition between w.ch closing and
	// waiter triggering isDone
	timer = time.NewTimer(shortTime)
	select {
	case <-isDone:
	case <-timer.C:
	}
	if !isWaitReturn.Load() {
		t.Error("Wait did not return on Done")
	}

}

// waiter tests WaitGroupCh.Wait()
func waiter(
	w *WaitGroupCh,
	isWaitReturn *atomic.Bool,
	isReady DoneLegacy,
	isDone chan struct{},
) {
	defer close(isDone)
	defer isWaitReturn.Store(true)

	isReady.Done()
	w.Wait()
}
