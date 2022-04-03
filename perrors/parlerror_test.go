/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"errors"
	"strings"
	"testing"
)

func TestParlErrorErrCh(t *testing.T) {
	chErr := make(chan error)
	closePanics := func() (didPanic bool) {
		defer func() {
			if v := recover(); v != nil {
				didPanic = true
			}
		}()

		close(chErr)
		return false
	}

	// does Shutdown close errCh?
	pe := NewParlError(chErr)
	pe.Shutdown()
	if !closePanics() {
		t.Error("pe.Shutdown did not close errCh")
	}

	expected := 1
	chErr = make(chan error)
	err := errors.New("message")

	// what happens on shutdown with blocking thread?
	pe = NewParlError(chErr)
	pe.AddError(err) // a thread is now blocked sending err on errCh
	pe.Shutdown()    // closes errCh
	list := ErrorList(pe.GetError())
	actualInt := len(list)

	if actualInt != expected {
		sList := make([]string, len(list))
		for i, e := range list {
			sList[i] = e.Error()
		}
		t.Errorf("Error count: %d expected: %d: [%s]", actualInt, expected, strings.Join(sList, ",\x20"))
	}

}

func TestParlError(t *testing.T) {
	e1 := errors.New("error1")
	e2 := errors.New("error2")

	var err error
	var actualString string
	var actualInt int

	// use of uninitialized instance
	var parlError ParlError
	err = parlError.GetError()
	if err != nil {
		t.Error("initial err was not nil")
	}

	// Error() when error is nil
	actualString = parlError.Error()
	if actualString != peNil {
		t.Errorf("bad Error() value for nil: %q expected: %q", actualString, peNil)
	}

	err = parlError.AddError(e1)
	if err != e1 {
		t.Error("First AdError bad return value")
	}
	parlError.AddErrorProc(e2)

	actualString = parlError.Error()
	if actualString != e1.Error() {
		t.Errorf("bad Error() value: %q expected: %q", actualString, e1.Error())
	}

	err = parlError.GetError()
	actualInt = len(ErrorList(err))
	if actualInt != 2 {
		t.Errorf("bad error encapsulation: %q expected: 2", actualInt)
	}
}

func TestParlErrorp(t *testing.T) {

	var err error

	// a nil parlError pointer causes panics
	// this will not usually happen, because pointers are not used
	// parlError nil situations
	var parlErrorp *ParlError

	// panic: runtime error: invalid memory address or nil pointer dereference
	//err = parlErrorp.GetError()
	_ = parlErrorp
	_ = err
}
