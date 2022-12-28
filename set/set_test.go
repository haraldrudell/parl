/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package set

import (
	"strings"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pslices"
)

func TestNewSet(t *testing.T) {
	value := 1
	name := "nname"
	notValue := 2
	notName := "?'2'"
	str := "int:1"
	messageDup := "duplicate set-element"

	var actual string

	interfaceSet := NewSet(pslices.ConvertSliceToInterface[
		Element[int],
		parl.Element[int],
	]([]Element[int]{{value, name}}))

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
	parl.RecoverInvocationPanic(func() {
		NewSet(pslices.ConvertSliceToInterface[
			Element[int],
			parl.Element[int],
		]([]Element[int]{
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
