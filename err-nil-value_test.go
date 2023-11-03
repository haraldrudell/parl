/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

func TestNilError(t *testing.T) {
	var argName = "xValue"
	var suffix = " cannot be nil"
	type iface interface{ NilValueError() }

	// get NilErr error value
	packFunc, err := newX(argName)
	if err == nil {
		t.Error("expected error missing")
		t.FailNow()
	}
	if packFunc == "" {
		t.Error("packFunc empty")
	}

	// error chain:
	// err: *errorglue.errorStack “parl.newX xValue cannot be nil” isErrNil: false
	// err: *parl.nilValue “parl.newX xValue cannot be nil” isErrNil: true
	var eList []string
	for e := err; e != nil; e = errors.Unwrap(e) {
		_, isErrNil := e.(iface)
		eList = append(eList, fmt.Sprintf("err: %T “%+[1]v” isErrNil: %t",
			e,
			isErrNil,
		))
	}
	t.Logf("error chain:\n%s", strings.Join(eList, "\n"))

	var expMessage = packFunc + "\x20" + argName + suffix

	// error should be ErrNil
	if !errors.Is(err, ErrNil) {

		_, errIs := err.(iface)
		var errNilErr error = ErrNil
		_, errNilIs := errNilErr.(iface)
		t.Errorf("err not ErrNil: err: %t ErrNil: %t", errIs, errNilIs)
	}
	if err.Error() != expMessage {
		t.Errorf("error message:\n%q exp\n%q",
			err.Error(),
			expMessage,
		)
	}

	// non-nill error should return false
	var err2 = perrors.New("x")
	if errors.Is(err2, ErrNil) {
		t.Error("err2 is ErrNil")
	}
}

// newX generates a NilError similar to a new function
func newX(argName string) (packFunc string, err error) {
	packFunc = pruntime.NewCodeLocation(0).PackFunc()
	err = NilError(argName)
	return
}
