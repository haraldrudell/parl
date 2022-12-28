/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"errors"
	"testing"

	"github.com/haraldrudell/parl"
)

func TestGoGroup(t *testing.T) {
	messageBad := "bad"
	errBad := errors.New(messageBad)

	var goGroup parl.GoGroup
	var goGroupImpl *GoGroup
	var g0 parl.Go
	var goError parl.GoError
	var goError2 parl.GoError
	var ok bool

	// instantiate a G1Group
	goGroup = NewGoGroup(context.Background())
	if goGroup == nil {
		t.Error("NewGoGroup nil")
		t.FailNow()
	}
	if goGroupImpl, ok = goGroup.(*GoGroup); !ok {
		t.Error("type assertion failed")
		t.FailNow()
	}

	// fail a thread
	g0 = goGroup.Go()
	if g0 == nil {
		t.Error("g1Group.Go nil")
		t.FailNow()
	}
	goGroupImpl.GoDone(g0, errBad)

	// check G1Error
	count := goGroupImpl.ch.Count()
	if count != 1 {
		t.Errorf("bad Ch Count: %d exp 1", count)
	}
	goError = <-goGroup.Ch()
	if goError == nil {
		t.Error("goError nil")
		t.FailNow()
	}
	if !errors.Is(goError.Err(), errBad) {
		t.Errorf("wrong error: %q %x exp %q %x", goError.Error(), goError, errBad.Error(), errBad)
	}
	if !goGroupImpl.isEnd() {
		t.Error("g1group did not terminate")
	}

	// instangiate g1 group
	goGroup = NewGoGroup(context.Background())
	goGroupImpl = goGroup.(*GoGroup)

	// ConsumeError
	g0 = goGroup.Go()
	goError2 = NewGoError(errBad, parl.GeNonFatal, g0)
	goGroupImpl.ConsumeError(goError2)
	if goGroupImpl.isEnd() {
		t.Error("g1group terminated")
	}

	//	String
	t.Log(goGroup.String())
	//t.Fail()
}
