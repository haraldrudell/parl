/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"github.com/haraldrudell/parl/pslices/pslib"
)

// SliceAwayAppend avoids allocations when a slice is
// sliced away from the beginning while being appended to at the end
//   - slicedAway: the slice of active values, sliced-away and appended to
//   - slice0: the original slicedAway
//   - values: values that should be appended to sliceAway
//   - noZeroOut [parl.NoZeroOut]: do not set unused element to zero-value.
//     Slices retaining values containing pointers in unused elements
//     is a temporary memory leak. Zero-out prevents this memory leak
//   - by storing the initial slice along with the slice-away slice,
//     the initial slice can be retrieved which may avoid allocations
//   - SliceAwayAppend takes pointer to slice so it can
//     update slicedAway and slice0
//   - There are three outcomes for a slice-away append:
//   - — 1 realloc: the result is larger than the underlying array: cap(*slice0)
//   - — 2 append: appending fits slicedAway capacity: cap(*slicedAway)
//   - — 3 copy: appending to SlicedAway fits the underlying array but
//     not slicedAway capacity
func SliceAwayAppend[T any](slicedAway, slice0 *[]T, values []T, noZeroOut ...pslib.ZeroOut) {

	// awaySlice is slicedAway prior to append
	var awaySlice = *slicedAway
	var awayLen = len(awaySlice)
	var awayCap = cap(awaySlice)
	// valueLength is number of elements on append complete
	var valueLength = awayLen + len(values)

	// case 2 append: appending fits slicedAway capacity as-is
	if valueLength <= awayCap {
		// increase the slice length
		awaySlice = awaySlice[:valueLength]
		// copy in values
		copy(awaySlice[awayLen:], values)
		// update source slice
		*slicedAway = awaySlice
		return // case 2 regular append return
	}

	// sliceZero is slice0 containing the full capacity from make
	var sliceZero = *slice0

	// case 1 realloc: the result is larger than the underlying array
	//	- update slice0
	if valueLength > cap(sliceZero) {
		// new slice0 based on slicedAway
		//	- re-alloc here
		awaySlice = append(awaySlice, values...)
		// update sliceAway and slice0
		*slicedAway = awaySlice
		*slice0 = awaySlice
		return // case 1 realloc return
	}
	// case 3 copy: appending to SlicedAway fits the underlying slice0 array but
	// not slicedAway capacity
	//	- slicedAway values need to be copied to
	//		the beginning of sliceZero

	// set new sliceZero length for copy
	if len(sliceZero) != valueLength {
		sliceZero = sliceZero[:valueLength]
	}

	// copy awaySlice to start of slice0
	copy(sliceZero, awaySlice)

	// append values
	copy(sliceZero[len(awaySlice):], values)
	// update source slice
	*slicedAway = sliceZero

	// check if zero-out should be carried out
	if len(noZeroOut) > 0 && noZeroOut[0] == pslib.NoZeroOut {
		return // case 3 copy, no zero-out return
	}

	// offset is how many elements slicedAway is sliced-away from slice0
	var offset, isValid = Offset(sliceZero, awaySlice)
	if !isValid {
		return // slice inconsistency return
	}

	// firstIndexToClear is at length of sliceZero
	//	- OK as low index for sliceZero slice-expression
	var firstIndexToClear = valueLength
	// lastIndexToClear is less than capacity of sliceZero
	//	- OK for high index for sliceZero slice-expression
	var lastIndexToClear = offset + len(awaySlice)
	if lastIndexToClear <= firstIndexToClear {
		return // no clear required return
	}
	clear(sliceZero[firstIndexToClear:lastIndexToClear])
}
