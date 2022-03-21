/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

import (
	"errors"
	"testing"
)

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
