/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"testing"

	"golang.org/x/exp/slices"
)

func TestSliceStringify(t *testing.T) {
	byts := []byte{0}
	exp := []string{"0"}

	var sList []string

	if sList = StringifySlice[int](nil); sList != nil {
		t.Error("nil non-nil")
	}

	if sList = StringifySlice(byts); slices.Compare(sList, exp) != 0 {
		t.Errorf("bad result: %v exp %v", sList, exp)
	}
}
