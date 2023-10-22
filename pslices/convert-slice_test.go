/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"strings"
	"testing"
)

func TestConvertSliceToInterface(t *testing.T) {
	type a int
	type b any

	// convert slice of a: int to slice of b:any
	interfaceSlice := ConvertSliceToInterface[
		a,
		b,
	]([]a{1})

	// slice length should be 1
	if len(interfaceSlice) != 1 {
		t.Errorf("bad ifSlice len: %d exp 1", len(interfaceSlice))
	}

	messageNI := "not implement"

	// panic conversion
	var err error
	func() {
		defer func() {
			err = recover().(error)
		}()

		ConvertSliceToInterface[
			b,
			a,
		]([]b{1})
	}()

	// should have anic not implement
	if err == nil || !strings.Contains(err.Error(), messageNI) {
		t.Errorf("ConvertSliceToInterface2 err: '%v' exp %q", err, messageNI)
	}
}
