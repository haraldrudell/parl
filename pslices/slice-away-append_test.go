/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"slices"
	"testing"
)

// There are three outcomes for a slice-away append:
//   - 1 realloc: the result is larger than the underlying array
//   - 2 append: appending fits slicedAway capacity
//   - 3 copy: appending to SlicedAway fits the underlying array but
//     not slicedAway capacity

func TestSliceAwayAppend_Realloc(t *testing.T) {
	//t.Error("logging on")
	var (
		// unsliced slice0
		slice00 = []int{1, 2, 3}
		// values to append 4 5 6
		values = []int{4, 5, 6}
		// slice-away index for 1-element slicedAway slice
		slicedAwayIndex = 1
		// expected slice0 slicedAway result: 2, 4, 5, 6
		expSlice0 = append([]int{slice00[slicedAwayIndex]}, values...)
	)

	var (
		offset  int
		isValid bool
	)

	var slice0 = slice00
	var slicedAway = slice0[slicedAwayIndex : slicedAwayIndex+1]
	// before: slice0: [1 2 3] slicedAway: [2] values: [4 5 6]
	t.Logf("before: slice0: %v slicedAway: %v values: %v", slice0, slicedAway, values)
	SliceAwayAppend(&slicedAway, &slice0, values)
	// after: slice0: [2 4 5 6] slicedAway: [2 4 5 6]
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

func TestSliceAwayAppend_Append(t *testing.T) {
	//t.Error("logging on")
	var (
		// a slice with all values known 1 2 3 4
		slice00 = []int{1, 2, 3, 4}
		// values to append 5
		values = []int{5}
		// slice-away index for 1-element slicedAway slice
		sliceAwayIndex = 1
		// 1 2 5 4
		expSlice0 = append(append(slices.Clone(slice00[:sliceAwayIndex+1]), values...), slice00[sliceAwayIndex+1+len(values):]...)
		// 2 5
		expSlicedAway = append([]int{slice00[sliceAwayIndex]}, values...)
	)

	var (
		offset  int
		isValid bool
	)

	var slice0 = slice00
	var slicedAway = slice0[sliceAwayIndex : sliceAwayIndex+1]
	// before: slice0: [1 2 3 4] slicedAway: [2] values: [5]
	t.Logf("before: slice0: %v slicedAway: %v values: %v", slice0, slicedAway, values)
	SliceAwayAppend(&slicedAway, &slice0, values)
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

func TestSliceAwayAppend_Copy(t *testing.T) {
	//t.Error("logging on")
	// requirements:
	//	- slicedAway and values should fit slice00
	//	- slicedAway should be sliced away so far that its capacity cannot
	//		fit values
	//	- the last element of slice00 should be untouched
	//	- the second-to-last element of slice00 should be zeroed out
	//	- slicedAway can be length 1
	//	- therefore, values must be at least length 2
	//	- values and slicedAway must be less than the second to last element of slicedAway
	//	- calc: 2 + 1 < len(slice00) - 2: slice00 is length 5
	var (
		// unsliced slice0
		slice00 = []int{1, 2, 3, 4, 5}
		// values to append 5
		values           = []int{6, 7}
		slicedAwayIndex0 = len(slice00) - 2
		slicedAwayIndex1 = len(slice00) - 1
		// 4, 6, 7, 0, 5
		expSlice0 = append(
			append(
				append(
					// slicedAway
					slices.Clone(slice00[slicedAwayIndex0:slicedAwayIndex1]),
					// values
					values...,
				),
				// zero-out
				0,
			),
			// 5
			slice00[slicedAwayIndex1:]...,
		)
		// 3 5
		expSliceAway = expSlice0[:slicedAwayIndex1-slicedAwayIndex0+len(values)]
	)

	var (
		offset  int
		isValid bool
	)

	// re-use slice should work
	var slice0 = slice00
	var slicedAway = slice0[slicedAwayIndex0:slicedAwayIndex1]
	// before: slice0: [1 2 3 4 5] slicedAway: [4] values: [6 7]
	t.Logf("before: slice0: %v slicedAway: %v values: %v", slice0, slicedAway, values)
	SliceAwayAppend(&slicedAway, &slice0, values)
	// after: slice0: [4 6 7 0 5] slicedAway: [4 6 7]
	t.Logf("after: slice0: %v slicedAway: %v", slice0, slicedAway)

	// slicedAway should match
	if !slices.Equal(slicedAway, expSliceAway) {
		t.Errorf("FAIL slicedAway %v exp %v", slicedAway, expSliceAway)
	}

	// slice0 should match
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
	//		slicedAway must be at the end of slice0
	//	- slicedAway and values should fit slice00
	//	- the last element of slice00 should be zeroed out
	//	- slicedAway can be length 1
	//	- for zero-out to happen, slice0 must be longer than
	//		slicedAway and value
	//	- len(slice00) = len(slicedAway) = len(value) + 1: 3
	var (
		// unsliced slice0
		slice00 = []int{1, 2, 3}
		// values to append 5
		value            = 4
		slicedAwayIndex0 = len(slice00) - 1
		slicedAwayIndex1 = len(slice00)
		// 3, 4, 0
		expSlice0 = append(
			append(
				// slicedAway
				slices.Clone(slice00[slicedAwayIndex0:slicedAwayIndex1]),
				// values
				value,
			),
			// zero-out
			0,
		)
		// 3 4
		expSliceAway = expSlice0[:slicedAwayIndex1-slicedAwayIndex0+1]
	)

	var (
		offset  int
		isValid bool
	)

	// re-use slice should work
	var slice0 = slice00
	var slicedAway = slice0[slicedAwayIndex0:slicedAwayIndex1]
	// before: slice0: [1 2 3] slicedAway: [3] values: 4
	t.Logf("before: slice0: %v slicedAway: %v values: %v", slice0, slicedAway, value)
	SliceAwayAppend1(&slicedAway, &slice0, value)
	// after: slice0: [3 4 0] slicedAway: [3 4]
	t.Logf("after: slice0: %v slicedAway: %v", slice0, slicedAway)

	// slicedAway should match
	if !slices.Equal(slicedAway, expSliceAway) {
		t.Errorf("FAIL slicedAway %v exp %v", slicedAway, expSliceAway)
	}

	// slice0 should match
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
		t.Error("FAIL isValid false")
	}
	if offset != sliceIndex {
		t.Errorf("FAIL bad offset %d exp %d", offset, sliceIndex)
	}
}
