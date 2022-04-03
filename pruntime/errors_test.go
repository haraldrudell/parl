/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"errors"
	"testing"
)

func TestIsRuntimeError(t *testing.T) {

	// test non-runtime.Error
	err := errors.New("err")
	actual := IsRuntimeError(err)
	if actual != nil {
		t.Error("IsRuntimeError true when expected false")
	}

	// test runtime.Error
	getRuntimeError := func() (err error) {
		defer func() {
			err = recover().(error)
		}()
		var ch chan struct{}
		close(ch)
		return
	}
	err = getRuntimeError()
	actual = IsRuntimeError(err)
	if actual == nil {
		t.Errorf("IsRuntimeError bad result: %v expected err: %v", actual, err)
	}
}

func TestIsSendOnClosedChannel(t *testing.T) {
	// test runtime.Error
	getRuntimeError := func() (err error) {
		defer func() {
			err = recover().(error)
		}()
		var ch chan struct{}
		close(ch)
		return
	}
	err := getRuntimeError()
	actualBool := IsSendOnClosedChannel(err)
	if actualBool {
		t.Error("IsSendOnClosedChannel true when expected false")
	}

	getCloseError := func() (err error) {
		defer func() {
			err = recover().(error)
		}()
		ch := make(chan struct{})
		close(ch)
		ch <- struct{}{}
		return
	}
	err = getCloseError()
	actualBool = IsSendOnClosedChannel(err)
	if !actualBool {
		t.Errorf("IsSendOnClosedChannel\n%q when expected true:\n%q", rtSendOnClosedChannel, err)
	}
}
