/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"slices"
	"testing"
)

func TestSliceAwayAppend1_Realloc(t *testing.T) {
	//t.Error("logging on")
	var (
		// unsliced slice0
		slice00 = []int{1}
		// values to append 4 5 6
		value = 4
		// expected slice0 slicedAway result: 2, 4, 5, 6
		expSlice0 = append(slices.Clone(slice00), value)
	)

	var (
		offset  int
		isValid bool
	)

	var slice0 = slice00
	var slicedAway = slice0
	// before: slice0: [1] slicedAway: [1] value: 4
	t.Logf("before: slice0: %v slicedAway: %v value: %v", slice0, slicedAway, value)
	SliceAwayAppend1(&slicedAway, &slice0, value)
	// after: slice0: [1 4] slicedAway: [1 4]
	t.Logf("after: slice0: %v slicedAway: %v", slice0, slicedAway)

	// slicedAway value should match
	if !slices.Equal(slicedAway, expSlice0) {
		t.Errorf("FAIL slicedAway %v exp %v", slicedAway, expSlice0)
	}

	// slice0 value should match
	if !slices.Equal(slice0, expSlice0) {
		t.Errorf("FAIL slice0 %v exp %v", slice0, expSlice0)
	}

	// slice0 and slicedAway should share underlying array
	offset, isValid = Offset(slice0, slicedAway)
	_ = offset
	if !isValid {
		t.Error("FAIL slice0 slicedAway not same underlying array")
	}

	// slice0 should be reallocated away from slice00
	offset, isValid = Offset(slice0, slice00)
	_ = offset
	if isValid {
		t.Error("FAIL slice0 slice00 is same underlying array")
	}
}

func TestSliceAwayAppend1_Append(t *testing.T) {
	//t.Error("logging on")
	var (
		// a slice with all values known 1 2 3 4
		slice00 = []int{1, 2, 3, 4}
		// values to append 5
		value = 5
		// slice-away index for 1-element slicedAway slice
		sliceAwayIndex = 1
		// 1 2 5 4
		expSlice0 = append(append(slices.Clone(slice00[:sliceAwayIndex+1]), value), slice00[sliceAwayIndex+2:]...)
		// 2 5
		expSlicedAway = append([]int{slice00[sliceAwayIndex]}, value)
	)

	var (
		offset  int
		isValid bool
	)

	var slice0 = slice00
	var slicedAway = slice0[sliceAwayIndex : sliceAwayIndex+1]
	// before: slice0: [1 2 3 4] slicedAway: [2] value: 5
	t.Logf("before: slice0: %v slicedAway: %v value: %v", slice0, slicedAway, value)
	SliceAwayAppend1(&slicedAway, &slice0, value)
	// after: slice0: [1 2 5 4] slicedAway: [2 5]
	t.Logf("after: slice0: %v slicedAway: %v", slice0, slicedAway)

	// slicedAway value should match
	if !slices.Equal(slicedAway, expSlicedAway) {
		t.Errorf("FAIL slicedAway %v exp %v", slicedAway, expSlicedAway)
	}

	// slice0 value should match
	if !slices.Equal(slice0, expSlice0) {
		t.Errorf("FAIL slice0 %v exp %v", slice0, expSlice0)
	}

	// slice0 and slicedAway should share underlying array
	offset, isValid = Offset(slice0, slicedAway)
	_ = offset
	if !isValid {
		t.Error("FAIL slice0 slicedAway not same underlying array")
	}

	// slice0 and slice00 should share underlying array
	offset, isValid = Offset(slice0, slice00)
	_ = offset
	if !isValid {
		t.Error("FAIL slice0 slice00 not same underlying array")
	}
}

func TestSliceAwayAppend1_Copy(t *testing.T) {
	//t.Error("logging on")
	// requirements:
	//	- to force a copy while appending a single element,
	//		slicedAway should be at the end of slice0
	//	- slicedAway and the appended value should fit slice0
	//	- the last element of slice0 should be zeroed out,
	//		therefore slice0 should be at least one longer than
	//		slicedAway and value
	//	- slicedAway can be length 1
	//	- slice0 should then be capacity 3 or more
	var (
		// slice00: [1, 2, 3]
		//	- slice00 points to the original slice0 underlying array
		slice00 = []int{1, 2, 3}
		// slice0 before: [1, 2, 3]
		//	- slice0 after: [3, 4, 0]
		slice0 = slice00
		// value to append: 4
		value = 4
		// slicedAway before: last index of slice0: [3] length 1
		slicedAway = slice0[2:3]
		// slice0 after: [3, 4, 0]
		expSlice0 = []int{slicedAway[0], value, 0}
		// slicedAway after: [3 4] length 2
		expSliceAway = []int{slicedAway[0], value}
	)

	var (
		offset  int
		isValid bool
	)

	// before: slice0: [1 2 3] slicedAway: [3] value: 4
	t.Logf("before: slice0: %v slicedAway: %v value: %v", slice0, slicedAway, value)

	SliceAwayAppend1(&slicedAway, &slice0, value)

	// after: slice0: [3 4 0] slicedAway: [3 4]
	t.Logf("after: slice0: %v slicedAway: %v", slice0, slicedAway)

	// slicedAway should match
	if !slices.Equal(slicedAway, expSliceAway) {
		t.Errorf("FAIL slicedAway\n%v exp\n%v",
			slicedAway, expSliceAway,
		)
	}

	// slice0 should match
	if !slices.Equal(slice0, expSlice0) {
		t.Errorf("FAIL slice0\n%v exp\n%v",
			slice0, expSlice0,
		)
	}

	// slice0 and slicedAway should share underlying array
	offset, isValid = Offset(slice0, slicedAway)
	_ = offset
	if !isValid {
		t.Error("FAIL slice0 slicedAway not same underlying array")
	}

	// slice0 and slice00 should share underlying array
	//	- otherwise, there has been allocation
	offset, isValid = Offset(slice0, slice00)
	_ = offset
	if !isValid {
		t.Error("FAIL slice0 slice00 not same underlying array")
	}
}
