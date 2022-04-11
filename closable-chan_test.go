/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"testing"
)

func TestCloser(t *testing.T) {

	// Closer can be invoked multiple times
	ch := make(chan struct{})
	cl := NewCloser(ch)
	ch2 := cl.Ch()
	if ch2 != ch {
		t.Errorf("NewCloser Ch bad")
	}
	if cl.IsClosed() {
		t.Errorf("NewCloser isClosed true")
	}
	// first close
	err, didClose := cl.Close()
	if err != nil {
		t.Errorf("cl.Close err %T %[1]v", err)
	}
	if !didClose {
		t.Errorf("cl.Close didClose false")
	}
	// second close
	err, didClose = cl.Close()
	if err != nil {
		t.Errorf("cl.Close err %T %[1]v", err)
	}
	if didClose {
		t.Errorf("cl.Close didClose true")
	}
	if !cl.IsClosed() {
		t.Errorf("NewCloser isClosed true")
	}

	// does not panic
	{
		ch := make(chan struct{})
		close(ch)
		cl := NewCloser(ch)
		var err2 error
		err, didClose := cl.Close(&err2)
		if err == nil {
			t.Errorf("cl.Close no err")
		}
		if err2 == nil {
			t.Errorf("cl.Close no err2")
		}
		if !didClose {
			t.Errorf("cl.Close didClose false")
		}
	}
}
