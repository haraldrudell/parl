/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslice

import (
	"strings"
	"testing"

	"github.com/haraldrudell/parl/recover"
)

func TestConvertSliceToInterface(t *testing.T) {
	type a int
	type b any
	interfaceSlice := ConvertSliceToInterface[
		a,
		b,
	]([]a{1})
	if len(interfaceSlice) != 1 {
		t.Errorf("bad ifSlice len: %d exp 1", len(interfaceSlice))
	}

	messageNI := "not implement"

	var err error
	recover.RecoverInvocationPanic(func() {
		ConvertSliceToInterface[
			b,
			a,
		]([]b{1})
		if len(interfaceSlice) != 1 {
			t.Errorf("bad ifSlice len: %d exp 1", len(interfaceSlice))
		}
	}, &err)
	if err == nil || !strings.Contains(err.Error(), messageNI) {
		t.Errorf("ConvertSliceToInterface2 err: '%v' exp %q", err, messageNI)
	}
}