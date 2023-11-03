/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"errors"
	"strconv"
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

func TestClose(t *testing.T) {
	// “nil pointer”
	//	- part of panic message:
	//	- “runtime error: invalid memory address or nil pointer dereference”
	var message = "nil pointer"
	// “x”
	var messageX = "x"
	// “y”
	var messageY = "y"

	var err error

	// a successful Close should not return error
	Close(newTestClosable(nil), &err)
	if err != nil {
		t.Errorf("Close err: %v", err)
	}

	// Close of nil io.Closable should return error, not panic
	err = nil
	Close(nil, &err)

	// err: panic detected in parl.Close:
	// “runtime error: invalid memory address or nil pointer dereference”
	// at parl.Close()-closer.go:40
	//t.Logf("err: %s", perrors.Short(err))
	//t.Fail()

	if err == nil || !strings.Contains(err.Error(), message) {
		t.Errorf("Close err: %v exp %q", err, message)
	}

	// a failed Close should append to err
	//	- err should become X with Y appended
	err = errors.New(messageX)
	Close(newTestClosable(errors.New(messageY)), &err)
	// errs is a list of associated errors, oldest first
	//	- first X, then Y
	var errs = perrors.ErrorList(err)
	if len(errs) != 2 || errs[0].Error() != messageX || !strings.HasSuffix(errs[1].Error(), messageY) {
		var quoteList = make([]string, len(errs))
		for i, err := range errs {
			quoteList[i] = strconv.Quote(err.Error())
		}
		// erss bad: ["x" "parl.Close y"]
		t.Errorf("erss bad: %v", quoteList)
	}
}

// testClosable is an [io.Closable] with configurable fail error
type testClosable struct{ err error }

// newTestClosable returns a closable whose Close always fails with err
func newTestClosable(err error) (closable *testClosable) { return &testClosable{err: err} }

// Close returns a configured error
func (tc *testClosable) Close() (err error) { return tc.err }
