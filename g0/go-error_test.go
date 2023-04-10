/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"errors"
	"fmt"
	"testing"

	"github.com/haraldrudell/parl"
)

func TestGoError(t *testing.T) {
	err := errors.New("x")
	errContext := parl.GePreDoneExit

	var goError parl.GoError
	var g0 parl.Go

	goError = NewGoError(err, errContext, g0)
	if goError == nil {
		t.Error("goError nil")
		t.FailNow()
	}
	if !errors.Is(goError.Err(), err) {
		t.Error("bad error value")
	}

	if goError.ErrContext() != errContext {
		t.Error("bad error context")
	}
	if goError.Error() != err.Error() {
		t.Error("bad err message")
	}
	if goError.Time().IsZero() {
		t.Error("bad error time")
	}
	if goError.Go() != g0 {
		t.Error("bad g0")
	}
	goError.IsThreadExit()
	goError.IsFatal()
	// "GePreDoneExit x g1ID::g0.TestNewG1Error-g1-error_test.go:18"
	t.Log(goError.String())

	goError = NewGoError(nil, errContext, g0)
	_ = goError.String()
}

func TestGoErrorString(t *testing.T) {
	var f = "is "

	var goErrorImpl = GoError{}
	var goErrorp = &goErrorImpl
	var goError parl.GoError = goErrorp

	var isStringer bool
	_, isStringer = goError.(fmt.Stringer)
	t.Logf("isStringer: %t", isStringer)

	var s = fmt.Sprintf(f+"%s", goError.String())
	t.Logf("resulting s: %q", s)
	if s == f {
		t.Error("s empty")
	}
	//t.Fail()
}
