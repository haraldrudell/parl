/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
	"testing"
)

func TestClosableChanClosing(t *testing.T) {

	// Closer can be invoked multiple times
	ch := make(chan struct{})
	cl := NewClosableChan(ch)
	ch2 := cl.Ch()
	if ch2 != ch {
		t.Errorf("NewCloser Ch bad")
	}
	if cl.IsClosed() {
		t.Errorf("NewCloser isClosed true")
	}
	// first close
	didClose, err := cl.Close()
	if err != nil {
		t.Errorf("cl.Close err %T %[1]v", err)
	}
	if !didClose {
		t.Errorf("cl.Close didClose false")
	}
	// second close
	didClose, err = cl.Close()
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
		cl := NewClosableChan(ch)
		var err2 error
		didClose, err := cl.Close(&err2)
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

func TestByValue(t *testing.T) {
	var ch ClosableChan[int]
	_ = &ch

	// A sync.Mutex field cannot be passed by value
	// func passes lock by value: github.com/haraldrudell/parl.ClosableChan[int] contains sync.Mutex
	/*
		f := func(c ClosableChan[int]) {}
		f(ch)
	*/

	// instead, pass by reference
	g := func(c *ClosableChan[int]) {}
	g(&ch)

	// sync.Once can be passed by value
	// 230120 sync.Once can no longer be passed by value
	//var o sync.Once
	//fo := func(p sync.Once) {}
	//fo(o)

	// sync.WaitGroup unlike previously by 1.19.3 not be passed by value
	//var wg sync.WaitGroup

	// func passes lock by value: sync.WaitGroup contains sync.noCopy
	//fw := func(g sync.WaitGroup) {}

	// call of fw copies lock value: sync.WaitGroup contains sync.noCopy
	//fw(wg)

	// *sync.Cond can be pased by value
	c := sync.NewCond(&sync.Mutex{})
	fc := func(d *sync.Cond) {}
	fc(c)

}
