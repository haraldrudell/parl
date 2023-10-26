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

func TestRecoverPanic(t *testing.T) {
	var didGood bool
	fnGood := func() {
		didGood = true
	}
	messagePanic := "fnPanic"
	errFnPanic := errors.New(messagePanic)
	fnPanic := func() {
		panic(errFnPanic)
	}
	messageErrp := "errp cannot be nil"

	var err error

	if RecoverInvocationPanic(fnGood, &err); err != nil {
		t.Errorf("RecoverPanic err: '%v'", err)
	}
	if !didGood {
		t.Error("RecoverPanic fnGood not invoked")
	}

	if RecoverInvocationPanic(fnPanic, &err); err == nil || !strings.Contains(err.Error(), messagePanic) {
		t.Errorf("RecoverPanic2 err: '%v'", err)
	}

	err = nil
	func() {
		defer func() {
			if v := recover(); v != nil {
				var ok bool
				if err, ok = v.(error); !ok {
					err = perrors.Errorf("panic-value not error; %T '%[1]v'", v)
				}
			}
		}()
		RecoverInvocationPanic(func() {}, nil)
	}()
	if err == nil || !strings.Contains(err.Error(), messageErrp) {
		t.Errorf("RecoverPanic3 bad error: '%v' exp %q", err, messageErrp)
	}
}
