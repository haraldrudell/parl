/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import (
	"testing"
)

func TestAwaitable(t *testing.T) {

	var ch AwaitableCh
	var isClosed, didClose bool

	// Ch() Close() IsClosed()
	var awaitable *Awaitable
	var reset = func() {
		awaitable = NewAwaitable()
	}

	// Ch is non-nil
	reset()
	ch = awaitable.Ch()
	if ch == nil {
		t.Error("Ch nil")
	}
	// Ch is open
	var isOpen bool
	select {
	case <-ch:
		isOpen = false
	default:
		isOpen = true
	}
	if !isOpen {
		t.Error("Ch closed")
	}
	// IsClosed is false
	isClosed = awaitable.IsClosed()
	if isClosed {
		t.Error("IsClosed true")
	}

	// Close works
	didClose = awaitable.Close()
	if !didClose {
		t.Errorf("didClose false")
	}
	// Ch is closed
	select {
	case <-ch:
		isOpen = false
	default:
		isOpen = true
	}
	if isOpen {
		t.Error("Ch open")
	}
	// IsClosed is true
	isClosed = awaitable.IsClosed()
	if !isClosed {
		t.Error("IsClosed false")
	}
}
