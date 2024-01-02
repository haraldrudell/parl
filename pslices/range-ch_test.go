/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"testing"
)

func TestRangeCh(t *testing.T) {
	var value1 = 1
	var value2 = 2

	var threadSafeSlice = NewThreadSafeSlice[int]()
	threadSafeSlice.Append(value1)
	threadSafeSlice.Append(value2)
	var ch <-chan int
	var actual int
	var ok, isCloseInvoked, isChClosed, exitChClosed bool
	var exitCh <-chan struct{}

	// Ch() Close() State()
	var rangeCh *RangeCh[int] = NewRangeCh(threadSafeSlice)

	// Ch() should return channel
	ch = rangeCh.Ch()
	if ch == nil {
		t.Error("Ch nil")
	}

	// State should be operational
	isCloseInvoked, isChClosed, exitCh = rangeCh.State()
	if isCloseInvoked {
		t.Error("isCloseInvoked true")
	}
	if isChClosed {
		t.Error("isChClosed true")
	}
	select {
	case <-exitCh:
		exitChClosed = true
	default:
		exitChClosed = false
	}
	if exitChClosed {
		t.Error("exitChClosed closed")
	}

	// ch should send first value
	// read first value
	if actual, ok = <-ch; !ok {
		t.Error("1: ok: false")
	}
	if actual != value1 {
		t.Errorf("Bad actual 1: %d exp %d", actual, value1)
	}

	// Close should terminate RangeCh
	rangeCh.Close()

	// ch should be closed
	if actual, ok = <-ch; ok {
		t.Error("2: ok: true")
	}
	_ = actual

	// State should be shut down
	isCloseInvoked, isChClosed, exitCh = rangeCh.State()
	if !isCloseInvoked {
		t.Error("isCloseInvoked false")
	}
	if !isChClosed {
		t.Error("isChClosed false")
	}
	select {
	case <-exitCh:
		exitChClosed = true
	default:
		exitChClosed = false
	}
	if !exitChClosed {
		t.Error("exitChClosed not closed")
	}
}
