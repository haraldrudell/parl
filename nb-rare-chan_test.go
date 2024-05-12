/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"testing"
	"time"
)

func TestNBRareChan(t *testing.T) {
	var value1, value2 = 1, 2
	const shortTime = time.Millisecond
	var expLength = 2

	var ch <-chan int
	var actual int
	var actualValues []int
	var emptyAwaitable AwaitableCh
	var isClose, ok bool
	var timer *time.Timer
	var timerC = func() (C <-chan time.Time) {
		if timer == nil {
			timer = time.NewTimer(shortTime)
		} else {
			timer.Stop()
			if len(timer.C) > 0 {
				<-timer.C
			}
			timer.Reset(shortTime)
		}
		C = timer.C
		return
	}

	// Ch Close IsClose PanicCh Send StopSend
	var nbChan *NBRareChan[int]
	var reset = func() {
		nbChan = &NBRareChan[int]{}
	}

	// Send one value should work
	//	- also tests Ch for returning non-nil value
	reset()
	nbChan.Send(value1)
	select {
	case actual = <-nbChan.Ch():
	case <-timerC():
		t.Fatal("nbChan.Ch timeout")
	}
	if actual != value1 {
		t.Errorf("Send: %d exp %d", actual, value1)
	}

	// Send two values should work
	reset()
	nbChan.Send(value1)
	nbChan.Send(value2)
	select {
	case actual = <-nbChan.Ch():
	case <-timerC():
		t.Fatal("nbChan.Ch timeout")
	}
	if actual != value1 {
		t.Errorf("Send: %d exp %d", actual, value1)
	}
	select {
	case actual = <-nbChan.Ch():
	case <-timerC():
		t.Fatal("nbChan.Ch timeout")
	}
	if actual != value2 {
		t.Errorf("Send: %d exp %d", actual, value2)
	}

	// StopSend should be effective
	reset()
	emptyAwaitable = nbChan.StopSend()
	_ = emptyAwaitable
	nbChan.Send(value1)
	select {
	case <-nbChan.Ch():
		t.Error("StopSend not disabling Send")
	case <-timerC():
	}

	// StopSend emptyAwaitable should work
	reset()
	nbChan.Send(value1)
	emptyAwaitable = nbChan.StopSend()
	if emptyAwaitable == nil {
		t.Fatal("StopSend emptyawaitable nil")
	}
	select {
	case <-emptyAwaitable:
		t.Error("StopSend emptyAwaitable closed before channel empty")
	default:
	}
	select {
	case <-nbChan.Ch():
	case <-timerC():
		t.Fatal("nbChan.Ch timeout")
	}
	select {
	case <-emptyAwaitable:
	default:
		t.Error("StopSend emptyAwaitable not closing on empty channel")
	}

	// Close should not hang
	reset()
	nbChan.Close(nil, nil)
	// Close should close underlying channel
	ch = nbChan.Ch()
	// if ch is closed:
	//	- channel receive will happen, not default
	//	- ok will be false
	ok = true
	select {
	case actual, ok = <-ch:
	default:
		t.Error("nbChanClose ch receive default")
	}
	if ok {
		t.Error("nbChanClose ch ok true")
	}

	// Close should return thread-value and queue
	reset()
	nbChan.Send(value1)
	nbChan.Send(value2)
	nbChan.Close(&actualValues, nil)
	if len(actualValues) != expLength {
		t.Fatalf("Close bad length %d exp %d", len(actualValues), expLength)
	}
	if actualValues[0] != value1 {
		t.Errorf("Close first value %d exp %d", actualValues[0], value1)
	}
	if actualValues[1] != value2 {
		t.Errorf("Close second value %d exp %d", actualValues[1], value2)
	}

	// IsClose should start false
	reset()
	isClose = nbChan.IsClose()
	if isClose {
		t.Error("IsClose started true")
	}
	nbChan.Close(nil, nil)
	// IsClose should be true after Close
	isClose = nbChan.IsClose()
	if !isClose {
		t.Error("IsClose ineffective")
	}

	// PanicCh should return non-nil emptyAwaitable
	reset()
	emptyAwaitable = nbChan.PanicCh()
	if emptyAwaitable == nil {
		t.Fatal("PanicCh nil emptyAwaitable")
	}
	// PanicCh emptyAwaitable should not be triggered
	select {
	case <-emptyAwaitable:
		t.Error("PanicCh emptyAwaitable triggered")
	default:
	}
}
