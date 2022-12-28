/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

// ConvertSliceToInterface converts a slice of a struct type to a slice of an interface type.
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
