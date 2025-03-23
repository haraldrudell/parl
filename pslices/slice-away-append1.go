/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import "github.com/haraldrudell/parl/pslices/pslib"

// SliceAwayAppend1 avoids allocations when a slice is
// sliced away from the beginning and appended to at the end
//   - sliceAway: the slice of active values, sliced away and appended to
//   - slice0: the original sliceAway
//   - value: the value that should be appended to sliceAway
//   - by storing the initial slice along with the slice-away slice,
//     the initial slice can be retrieved which may avoid allocations
//   - SliceAwayAppend avoid such allocations based on two pointers to slice
func SliceAwayAppend1[T any](slicedAway, slice0 *[]T, value T, noZeroOut ...pslib.ZeroOut) {

	// awaySlice is slicedAway prior to append
	var awaySlice = *slicedAway
	// valueLength is number of elements on append complete
	var valueLength = len(awaySlice) + 1

	// 2 append: appending fits slicedAway capacity
	if valueLength <= cap(awaySlice) {
		*slicedAway = append(awaySlice, value)
		return
	}

	// sliceZero is slice0 length and capacity prior to append
	var sliceZero = *slice0

	// case 1 realloc: the result is larger than the underlying array
	//	- update slice0
	if valueLength > cap(sliceZero) {
		// new slice0 based on slicedAway
		awaySlice = append(awaySlice, value)
		// update sliceAway and slice0
		*slicedAway = awaySlice
		*slice0 = awaySlice
		return // case 1 realloc return
	}
	// 3 copy: appending to SlicedAway fits the underlying array but
	// not slicedAway capacity
	//	- slicedAway values need to be copied to
	//		the beginning of sliceZero

	// set new sliceZero length for copy
	if len(sliceZero) != valueLength {
		sliceZero = sliceZero[:valueLength]
	}

	// copy awaySlice to start of slice0
	copy(sliceZero, awaySlice)

	// append value
	sliceZero[valueLength-1] = value
	*slicedAway = sliceZero

	// is zero-out disabled?
	if len(noZeroOut) > 0 && noZeroOut[0] == pslib.NoZeroOut {
		return // case 3 copy, no zero-out return
	}

	// offset is how many elements slicedAway is sliced-away from slice0
	var offset, isValid = Offset(sliceZero, awaySlice)
	if !isValid || offset < 2 {
		return // slice inconsistency or no zero-out return
	}
	clear(sliceZero[valueLength : valueLength+1])
}
