/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import (
	"testing"
)

func TestCyclicAwaitable(t *testing.T) {

	var ch, ch2 AwaitableCh
	var isClosed, didClose, didOpen bool

	// Ch() Close() IsClosed() Open()
	var cyclicAwaitable CyclicAwaitable
	var reset = func() {
		cyclicAwaitable = CyclicAwaitable{}
	}

	// Ch should return open channel
	reset()
	ch = cyclicAwaitable.Ch()
	if ch == nil {
		t.Errorf("ch nil")
	}
	select {
	case <-ch:
		isClosed = true
	default:
		isClosed = false
	}
	if isClosed {
		t.Errorf("ch closed")
	}

	// IsClosed should be false
	reset()
	isClosed = cyclicAwaitable.IsClosed()
	if isClosed {
		t.Errorf("IsClosed() true")
	}

	// Open should open
	reset()
	didClose = cyclicAwaitable.Close()
	_ = didClose
	didOpen, ch = cyclicAwaitable.Open()
	if !didOpen {
		t.Errorf("didOpen false")
	}
	if ch == nil {
		t.Errorf("ch nil")
	}
	select {
	case <-ch:
		isClosed = true
	default:
		isClosed = false
	}
	if isClosed {
		t.Errorf("ch not open")
	}
	isClosed = cyclicAwaitable.IsClosed()
	if isClosed {
		t.Errorf("IsClosed() true")
	}
	didOpen, ch2 = cyclicAwaitable.Open()
	if didOpen {
		t.Errorf("didOpen true")
	}
	if ch2 != ch {
		t.Errorf("not same channel")
	}

	// Close should close
	reset()
	didClose = cyclicAwaitable.Close()
	if !didClose {
		t.Errorf("didClose false")
	}
	isClosed = cyclicAwaitable.IsClosed()
	if !isClosed {
		t.Errorf("IsClosed() false")
	}
	ch = cyclicAwaitable.Ch()
	select {
	case <-ch:
		isClosed = true
	default:
		isClosed = false
	}
	if !isClosed {
		t.Errorf("ch not closed")
	}
	didClose = cyclicAwaitable.Close()
	if didClose {
		t.Errorf("didClose true")
	}
}
