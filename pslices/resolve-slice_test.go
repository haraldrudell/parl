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

func TestResolveSlice(t *testing.T) {
	v1 := 1
	v1p := &v1
	slic := []*int{v1p}
	slicResolved := []int{v1}

	var actual []int

	ResolveSlice[int](nil)
	actual = ResolveSlice(slic)
	if slices.Compare(actual, slicResolved) != 0 {
		t.Errorf("bad slice: %v exp %v", actual, slicResolved)
	}
}
