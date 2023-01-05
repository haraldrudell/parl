/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package set

import (
	"strings"
	"testing"

	"github.com/haraldrudell/parl/pslice"
	"github.com/haraldrudell/parl/test"
)

func TestNewSet(t *testing.T) {
	value := 1
	name := "nname"
	notValue := 2
	notName := "?'2'"
	str := "int:1"
	messageDup := "duplicate set-element"

	var actual string

	interfaceSet := NewSet(pslice.ConvertSliceToInterface[
		SetElement[int],
		Element[int],
	]([]SetElement[int]{{value, name}}))

	if interfaceSet == nil {
		t.Error("NewSet nil")
		t.FailNow()
	}
	if actual = interfaceSet.StringT(value); actual != name {
		t.Errorf("StringT %q exp %q", actual, name)
	}
	if actual = interfaceSet.StringT(notValue); actual != notName {
		t.Errorf("StringT2 %q exp %q", actual, notName)
	}
	if actual = interfaceSet.String(); actual != str {
		t.Errorf("String %q exp %q", actual, str)
	}

	var err error
	test.RecoverInvocationPanic(func() {
		NewSet(pslice.ConvertSliceToInterface[
			SetElement[int],
			Element[int],
		]([]SetElement[int]{
			{value, name},
			{value, name},
		}))
	}, &err)
	if err == nil {
		t.Error("set duplicate element missing error")
	} else if !strings.Contains(err.Error(), messageDup) {
		t.Errorf("NewSet2 err: %q exp %q", err.Error(), messageDup)
	}
}
