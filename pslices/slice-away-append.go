/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"unsafe"
)

// [SliceAwayAppend] [SliceAwayAppend1] do not zero-out obsolete slice elements
//   - [SetLength] noZero
const NoZeroOut = true

// [SliceAwayAppend] [SliceAwayAppend1] do zero-out obsolete slice elements
//   - [SetLength] noZero
const DoZeroOut = false

// SliceAwayAppend avoids allocations when a slice is
// sliced away from the beginning while being appended to at the end
//   - sliceAway: the slice of active values, sliced away and appended to
//   - slice0: the original sliceAway
//   - values: values that should be appended to sliceAway
//   - noZeroOut NoZeroOut: do not set unused element to zero-value.
//     Slices retaining values containing pointers in unused elements
//     is a temporary memory leak. Zero-out prevents this memory leak
//   - by storing the initial slice along with the slice-away slice,
//     the initial slice can be retrieved which may avoid allocations
//   - SliceAwayAppend takes pointer to slice so it can
//     update slicedAway and slice0
//   - There are three outcomes for a slice-away append:
//   - — 1 realloc: the result is larger than the underlying array
//   - — 2 append: appending fits slicedAway capacity
//   - — 3 copy: appending to SlicedAway fits the underlying array but
//     not slicedAway capacity
func SliceAwayAppend[T any](slicedAway, slice0 *[]T, values []T, noZeroOut ...bool) {
	// awaySlice is slicedAway prior to append
	var awaySlice = *slicedAway
	// sliceZero is slice0 pointer and capacity from make
	var sliceZero = *slice0

	// useRegularAppend indicates that a normal append
	// should be used
	//	- true for case 1 realloc and case 2 append
	//	- false for case 3 copy
	var valueLength = len(awaySlice) + len(values)

	// 2 append: appending fits slicedAway capacity
	if valueLength <= cap(awaySlice) {
		*slicedAway = append(awaySlice, values...)
		return
	}

	// 1 realloc: the result is larger than the underlying array
	if valueLength > cap(sliceZero) {
		awaySlice = append(awaySlice, values...)
		// update sliceAway and slice0
		*slicedAway = awaySlice
		*slice0 = awaySlice
		return
	}
	// 3 copy: appending to SlicedAway fits the underlying array but
	// not slicedAway capacity
	//	- slicedAway values need to be copied to
	//		the beginning of sliceZero

	// re-use slice0
	//	- cannot arbitrary set length to do copy
	//	- therefore, set length to zero and double-append
	if len(sliceZero) > 0 {
		sliceZero = sliceZero[:0]
	}
	// newSlice is at beginning of array and has the new aggregate length
	//	- appends do not cause allocation
	var newSlice = append(append(sliceZero, awaySlice...), values...)
	*slicedAway = newSlice

	// zero-out if not disabled
	if len(noZeroOut) > 0 && noZeroOut[0] {
		return // no zero-out
	}
	zeroOut(newSlice, awaySlice)
}

// SliceAwayAppend1 avoids allocations when a slice is
// sliced away from the beginning and appended to at the end
//   - sliceAway: the slice of active values, sliced away and appended to
//   - slice0: the original sliceAway
//   - value: the value that should be appended to sliceAway
//   - by storing the initial slice along with the slice-away slice,
//     the initial slice can be retrieved which may avoid allocations
//   - SliceAwayAppend avoid such allocations based on two pointers to slice
func SliceAwayAppend1[T any](slicedAway, slice0 *[]T, value T, noZeroOut ...bool) {
	// awaySlice is slicedAway prior to append
	var awaySlice = *slicedAway
	// sliceZero is slice0 length and capacity prior to append
	var sliceZero = *slice0

	// useRegularAppend indicates that a normal append
	// should be used
	//	- true for case 1 realloc and case 2 append
	//	- false for case 3 copy
	var valueLength = len(awaySlice) + 1

	// 2 append: appending fits slicedAway capacity
	if valueLength <= cap(awaySlice) {
		*slicedAway = append(awaySlice, value)
		return
	}

	// 1 realloc: the result is larger than the underlying array
	if valueLength > cap(sliceZero) {
		awaySlice = append(awaySlice, value)
		// update sliceAway and slice0
		*slicedAway = awaySlice
		*slice0 = awaySlice
		return
	}
	// 3 copy: appending to SlicedAway fits the underlying array but
	// not slicedAway capacity
	//	- slicedAway values need to be copied to
	//		the beginning of sliceZero

	// re-use slice0
	//	- cannot arbitrary set length to do copy
	//	- therefore, set length to zero and double-append
	if len(sliceZero) > 0 {
		sliceZero = sliceZero[:0]
	}
	// newSlice is at beginning of array and has the new aggregate length
	//	- appends do not cause allocation
	var newSlice = append(append(sliceZero, awaySlice...), value)
	*slicedAway = newSlice

	// is zero-out disabled?
	if len(noZeroOut) > 0 && noZeroOut[0] {
		return // no zero-out
	}
	zeroOut(newSlice, awaySlice)
}

// zero-out emptied values
//   - newSlice begins at the beginning of the underlying array and
//     contains all the newly appended aggregate data
//   - slicedAway is the previously slicedAway slice containing
//     data from before the append
//   - elements to zero out are those at end of slicedAway that
//     are not part of newSlice
func zeroOut[T any](newSlice, slicedAway []T) {

	// offset his how many elements slicedAway is off from newSlice
	var offset, isValid = Offset(newSlice, slicedAway)
	if !isValid {
		return // some issue with slices
	}

	// number of elements to zero out at end of slicedAway
	var elementCount int
	// the first index not used in newSlice now
	var newEnd = len(newSlice)
	// the first index not used before in newSlice
	var oldEnd = offset + len(slicedAway)
	// newEnd is greater or equal, there are no element to zero out
	if newEnd >= oldEnd {
		return // no elements to zero-out
	}
	// number of elements to zero out at the end of slicedAway
	elementCount = oldEnd - newEnd
	// limit elementCount to length of slicedAway
	if elementCount > len(slicedAway) {
		elementCount = len(slicedAway)
	}

	// zero out elementCount elements at the end of slicedAway
	var zeroValue T
	for index := len(slicedAway) - elementCount; index < len(slicedAway); index++ {
		slicedAway[index] = zeroValue
	}
}

// Offset calculates how many items a slice-away slice is off
// from the initial slice
//   - slice0 is a slice containing the beginning of the underlying array
//   - slicedAway is a slice that has been sliced-off at the beginning
func Offset[T any](slice0, slicedAway []T) (offset int, isValid bool) {

	// verify that operation is possible
	//	- slice0 slicedAway cannot be nil
	if slice0 == nil || slicedAway == nil {
		return // operation failed
	}

	// a pointer to a slice points to 3 values:
	//	- pointer to underlying array: pointer to array of elements
	//	- current slice length: int
	//	- current slice capacity: int
	//	- by casting uintptr to uint, pointer arithmetic becomes possible

	// slice0p is the uint memory address of the underlying array of slice0
	var slice0p = *((*uint)(unsafe.Pointer(&slice0)))
	// slicedAwayp is the uint memory address of the underlying array of slicedAway
	var slicedAwayp = *((*uint)(unsafe.Pointer(&slicedAway)))

	// // slice0p 0x14000016160
	// fmt.Fprintf(os.Stderr, "slice0p 0x%x\n", slice0p)
	// // slicedAwayp 0x14000016170
	// fmt.Fprintf(os.Stderr, "slicedAwayp 0x%x\n", slicedAwayp)

	// slicedAway[0]: 3
	// fmt.Fprintf(os.Stderr, "slicedAway[0]: %d\n",
	// 	**(**int)(unsafe.Pointer(&slicedAway)),
	// )

	// ensure slicedAway is from the same underlying array as slice0
	if slicedAwayp < slice0p {
		return // not same array
	}

	// determine element size in bytes
	var a [2]T
	var elementSize = 0 + //
		(uint)((uintptr)(unsafe.Pointer(&a[1]))) -
		(uint)((uintptr)(unsafe.Pointer(&a[0])))

		// int 64-bit elementSize 0x8
	//fmt.Fprintf(os.Stderr, "int 64-bit elementSize 0x%x\n", elementSize)

	// offset is how many elements have been sliced off from slice0
	//	- not negative
	offset = int((slicedAwayp - slice0p) / elementSize)
	isValid =
		// slicedAwayp - slice0p must be even divisible by element size
		slice0p+uint(offset)*elementSize == slicedAwayp &&
			// slicedAway and offset cannot beyond the end of slice0
			offset+len(slicedAway) <= cap(slice0)

	return
}
