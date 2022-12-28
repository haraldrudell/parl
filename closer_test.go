/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"errors"
	"strings"
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

func TestCloser(t *testing.T) {
	message := "close of closed channel"
	messageNil := "close of nil channel"

	var ch chan struct{}
	var err error
	var c chan struct{}

	ch = make(chan struct{})
	Closer(ch, &err)
	if err != nil {
		t.Errorf("Closer err: %v exp nil", err)
	}

	err = nil
	Closer(ch, &err)
	if err == nil || !strings.Contains(err.Error(), message) {
		t.Errorf("Closer closed err: %v exp %q", err, message)
	}

	ch = c
	err = nil
	Closer(ch, &err)
	if err == nil || !strings.Contains(err.Error(), messageNil) {
		t.Errorf("Closer nil err: %v exp %q", err, messageNil)
	}
}

func TestCloserSend(t *testing.T) {
	message := "close of closed channel"
	messageNil := "close of nil channel"

	var ch chan<- struct{}
	var err error
	var c chan<- struct{}

	ch = make(chan<- struct{})
	CloserSend(ch, &err)
	if err != nil {
		t.Errorf("Closer err: %v exp nil", err)
	}

	err = nil
	CloserSend(ch, &err)
	if err == nil || !strings.Contains(err.Error(), message) {
		t.Errorf("Closer closed err: %v exp %q", err, message)
	}

	ch = c
	err = nil
	CloserSend(ch, &err)
	if err == nil || !strings.Contains(err.Error(), messageNil) {
		t.Errorf("Closer nil err: %v exp %q", err, messageNil)
	}
}

type testClosable struct {
	err error
}

func (tc *testClosable) Close() (err error) { return tc.err }

func TestClose(t *testing.T) {
	message := "nil pointer"
	err1 := "x"
	err2 := "y"

	var err error

	Close(&testClosable{}, &err)
	if err != nil {
		t.Errorf("Close err: %v", err)
	}

	err = nil
	Close(nil, &err)
	if err == nil || !strings.Contains(err.Error(), message) {
		t.Errorf("Close err: %v exp %q", err, message)
	}

	err = errors.New(err1)
	Close(&testClosable{err: errors.New(err2)}, &err)

	errs := perrors.ErrorList(err)
	if len(errs) != 2 || errs[0].Error() != err1 || errs[1].Error() != err2 {
		t.Errorf("erss bad: %v", errs)
	}
}
