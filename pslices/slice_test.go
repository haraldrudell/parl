/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import "testing"

func TestSlice(t *testing.T) {
	v1 := 2

	var slice Slice[int]
	var actual int

	slice = Slice[int]{list: []int{v1}}

	if actual = slice.Length(); actual != 1 {
		t.Errorf("Length not 1: %d", actual)
	}

	if actual = slice.Element(0); actual != v1 {
		t.Errorf("Element not %d: %d", v1, actual)
	}
	if actual = slice.Element(1); actual != 0 {
		t.Errorf("Element not 0: %d", actual)
	}

	slice.Clear()
	slice.List()
}
