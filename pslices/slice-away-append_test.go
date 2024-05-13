/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"slices"
	"testing"
)

func TestSliceAwayAppend_ReUse(t *testing.T) {
	//t.Error("logging on")
	// a slice with all values known
	var slice = []int{1, 2, 3, 4}
	// values to append
	var values = []int{5}
	// slice-away index for 1-element slicedAway slice
	var sliceAwayIndex = 2
	var expSlice = []int{3, 5, 0, 4}
	var expSliceAway = expSlice[:2]

	// re-use slice should work
	var slice0 = slice[:0]
	// index 2:3: slice [3]
	var sliceAway = slice[sliceAwayIndex : sliceAwayIndex+1]
	SliceAwayAppend(&sliceAway, &slice0, values)
	// slice: [3 5 0 4] sliceAway: [3 5]
	t.Logf("slice: %v sliceAway: %v", slice, sliceAway)
	if !slices.Equal(sliceAway, expSliceAway) {
		t.Errorf("sliceAway %v exp %v", sliceAway, expSliceAway)
	}
	if !slices.Equal(slice, expSlice) {
		t.Errorf("slice %v exp %v", slice, expSlice)
	}
}

func TestSliceAwayAppend_Append(t *testing.T) {
	//t.Error("logging on")
	// a slice with all values known
	var slice = []int{1, 2}
	// values to append
	var values = []int{3, 4}
	// slice-away index for 1-element slicedAway slice
	var sliceAwayIndex = 1
	var expSlice = slices.Clone(slice)
	var expSliceAway = []int{2, 3, 4}

	// regular append should work
	var slice0 = slice[:0]
	// index 1:2: slice [2]
	var sliceAway = slice[sliceAwayIndex : sliceAwayIndex+1]
	SliceAwayAppend(&sliceAway, &slice0, values)
	// slice: [1 2] sliceAway: [2 3 4]
	//	- sliceAway is a newly allocated slice
	t.Logf("slice: %v sliceAway: %v", slice, sliceAway)
	if !slices.Equal(sliceAway, expSliceAway) {
		t.Errorf("sliceAway %v exp %v", sliceAway, expSliceAway)
	}
	if !slices.Equal(slice, expSlice) {
		t.Errorf("slice %v exp %v", slice, expSlice)
	}
}

func TestSliceAwayAppend1_ReUse(t *testing.T) {
	//t.Error("logging on")
	var slice = []int{1, 2, 3, 4}
	var value = 5
	var sliceAwayIndex = 2
	var expSlice = []int{3, 5, 0, 4}
	var expSliceAway = []int{3, 5}

	var slice0 = slice[:0]
	// index 2:3: slice [3]
	var sliceAway = slice[sliceAwayIndex : sliceAwayIndex+1]
	SliceAwayAppend1(&sliceAway, &slice0, value)
	// slice: [3 5 0 4] sliceAway: [3 5]
	t.Logf("slice: %v sliceAway: %v", slice, sliceAway)
	if !slices.Equal(sliceAway, expSliceAway) {
		t.Errorf("sliceAway %v exp %v", sliceAway, expSliceAway)
	}
	if !slices.Equal(slice, expSlice) {
		t.Errorf("slice %v exp %v", slice, expSlice)
	}
}

func TestSliceAwayAppend1_Append(t *testing.T) {
	//t.Error("logging on")
	var slice = []int{1, 2}
	var value = 5
	var expSlice = slices.Clone(slice)
	var expSliceAway = append(slices.Clone(slice), value)

	var slice0 = slice[:0]
	var sliceAway = slice
	SliceAwayAppend1(&sliceAway, &slice0, value)
	// slice: [1 2] sliceAway: [1 2 5]
	t.Logf("slice: %v sliceAway: %v", slice, sliceAway)
	if !slices.Equal(sliceAway, expSliceAway) {
		t.Errorf("sliceAway %v exp %v", sliceAway, expSliceAway)
	}
	if !slices.Equal(slice, expSlice) {
		t.Errorf("slice %v exp %v", slice, expSlice)
	}
}

func TestOffset(t *testing.T) {
	//t.Error("logging on")
	var size = 3
	var sliceIndex = 2

	var offset int
	var isValid bool

	var slice0 = make([]int, size)
	var slicedAway = slice0[sliceIndex:]
	offset, isValid = Offset(slice0, slicedAway)
	// offset: 2 isValid: true
	t.Logf("offset: %d isValid: %t", offset, isValid)
	if !isValid {
		t.Error("isValid false")
	}
	if offset != sliceIndex {
		t.Errorf("bad offset %d exp %d", offset, sliceIndex)
	}
}
