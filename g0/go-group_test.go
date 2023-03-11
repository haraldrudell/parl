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
	var goImpl *Go
	var goError parl.GoError
	var goError2 parl.GoError
	var ok bool
	var count int
	var subGo parl.SubGo
	var subGroup parl.SubGroup

	// g0.NewGoGroup returns *g0.GoGroup: NewGoGroup()
	goGroup = NewGoGroup(context.Background())
	if goGroupImpl, ok = goGroup.(*GoGroup); !ok {
		t.Error("NewGoGroup did not return *g0.GoGroup")
		t.FailNow()
	}

	// fail thread exit: Go() GoDone() Ch() Threads() NamedThreads() IsEnd()
	g0 = goGroup.Go()
	if goImpl, ok = g0.(*Go); !ok {
		t.Error("GoGroup.Go() did not return *g0.Go")
		t.FailNow()
	}
	_ = goImpl
	g0.Register("x")
	if len(goGroup.Threads()) != 1 {
		t.Error("goGroup.Threads length bad")
	}
	if len(goGroup.NamedThreads()) != 1 {
		t.Error("goGroup.NamedThreads length bad")
	}
	goGroupImpl.GoDone(g0, errBad)
	// verify that error channel has error
	if count = goGroupImpl.ch.Count(); count != 1 {
		t.Errorf("bad Ch Count: %d exp 1", count)
	}
	goError, ok = <-goGroup.Ch()
	if !ok {
		t.Error("goGroup.Ch closed")
	}
	if goError == nil {
		t.Error("goError nil")
		t.FailNow()
	}
	if !errors.Is(goError.Err(), errBad) {
		t.Errorf("wrong error: %q %x exp %q %x", goError.Error(), goError, errBad.Error(), errBad)
	}
	// verify GoGroup termination
	_, ok = <-goGroup.Ch()
	if ok {
		t.Error("goGroup.Ch did not close")
	}
	if !goGroupImpl.isEnd() {
		t.Error("goGroup did not terminate")
	}

	// ConsumeError() EnableTermination()
	goGroup = NewGoGroup(context.Background())
	goGroupImpl = goGroup.(*GoGroup)
	g0 = goGroup.Go()
	goError2 = NewGoError(errBad, parl.GeNonFatal, g0)
	goGroupImpl.ConsumeError(goError2)
	var goErrorActual = <-goGroup.Ch() // GoError from ConsumeError
	if goErrorActual != goError2 {
		t.Error("GoGroup.ConsumeError failed")
	}
	if goGroupImpl.isEnd() {
		t.Error("1 GoGroup terminated")
	}
	goGroup.EnableTermination(false)
	goGroupImpl.GoDone(g0, nil)
	<-goGroup.Ch() // GoError from GoDone
	if goGroupImpl.isEnd() {
		t.Error("2 GoGroup terminated")
	}
	goGroup.EnableTermination(true)
	<-goGroup.Ch() // allow error channel to close
	if !goGroupImpl.isEnd() {
		t.Error("GoGroup did not terminate")
	}

	// SubGo() SubGroup()
	goGroup = NewGoGroup(context.Background())
	subGo = goGroup.SubGo()
	goGroupImpl = subGo.(*GoGroup)
	if goGroupImpl.isSubGroup.IsTrue() {
		t.Error("SubGo returned SubGroup")
	}
	if goGroupImpl.hasErrorChannel.IsTrue() {
		t.Error("SubGo has error channel")
	}
	subGroup = goGroup.SubGroup()
	goGroupImpl = subGroup.(*GoGroup)
	if goGroupImpl.isSubGroup.IsFalse() {
		t.Error("SubGroup did not return SubGroup")
	}
	if goGroupImpl.hasErrorChannel.IsFalse() {
		t.Error("SubGroup does not have error channel")
	}

	// Add() UpdateThread() FirstFatal()

	//	String
	t.Log(goGroup.String())
	//t.Fail()
}
