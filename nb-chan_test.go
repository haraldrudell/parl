/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"testing"
)

func TestNewNBChan(t *testing.T) {
	var nbChan NBChan[int]

	iE := 1
	iF := 2
	countE := 2
	nbChan.Send(iE)
	nbChan.Send(iF)
	count := nbChan.Count()
	if count != countE {
		t.Errorf("nbChan bad count: %d exp %d", count, countE)
	}
	i, ok := <-nbChan.Ch()
	if !ok {
		t.Errorf("<-nbChan.Ch: ok false")
	}
	if i != iE {
		t.Errorf("nbChan bad receive: %d exp %d", i, iE)
	}
	if nbChan.IsClosed() {
		t.Errorf("nbChan.IsClosed: true")
	}
	didClose, err := nbChan.CloseNow()
	if err != nil {
		t.Errorf("nbChan.Close err: %v", err)
	}
	if !didClose {
		t.Errorf("nbChan.didClosed: bad")
	}
}
