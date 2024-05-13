/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"unsafe"
)

// [SliceAwayAppend] [SliceAwayAppend1] do not zero-out obsolete slice elements
const NoZeroOut = true

// SliceAwayAppend avoids allocations when a slice is
// sliced away from the beginning and appended to at the end
//   - sliceAway: the slice of active values, sliced away and appended to
//   - slice0: the original sliceAway
//   - values: values that should be appended to sliceAway
//   - by storing the initial slice along with the slice-away slice,
//     the initial slice can be retrieved which may avoid allocations
//   - SliceAwayAppend avoid such allocations based on two pointers to slice
func SliceAwayAppend[T any](slicedAway, slice0 *[]T, values []T, noZeroOut ...bool) {
	// awaySlice is slicedAway prior to append
	var awaySlice = *slicedAway
	// sliceZero is slice0 length and capacity prior to append
	var sliceZero = *slice0

	// if insufficient capacity, use regular append
	if len(awaySlice)+len(values) > cap(sliceZero) {
		awaySlice = append(awaySlice, values...)
		// update sliceAway and slice0
		*slicedAway = awaySlice
		*slice0 = awaySlice
		return
	}

	// re-use slice0
	//	- cannot arbitrary set length to do copy
	//	- therefore, set length to zero and double-append
	if len(sliceZero) > 0 {
		sliceZero = sliceZero[:0]
	}
	// newSlice is at beginning of array and has the new aggregate length
	var newSlice = append(append(sliceZero, awaySlice...), values...)
	*slicedAway = newSlice

	// is zero-out disabled?
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
func SliceAwayAppend1[T any](sliceAway, slice0 *[]T, value T, noZeroOut ...bool) {
	// awaySlice is slicedAway prior to append
	var awaySlice = *sliceAway
	// sliceZero is slice0 length and capacity prior to append
	var sliceZero = *slice0

	// if insufficient capacity, use regular append
	//	- no elements to zero out
	if len(awaySlice)+1 > cap(sliceZero) {
		awaySlice = append(awaySlice, value)
		// update sliceAway and slice0
		*sliceAway = awaySlice
		*slice0 = awaySlice
		return
	}

	// re-use slice0
	//	- cannot arbitrary set length to do copy
	//	- therefore, set length to zero and double-append
	if len(sliceZero) > 0 {
		sliceZero = sliceZero[:0]
	}
	var newSlice = append(append(sliceZero, awaySlice...), value)
	*sliceAway = newSlice

	// is zero-out disabled?
	if len(noZeroOut) > 0 && noZeroOut[0] {
		return // no zero-out
	}
	zeroOut(newSlice, awaySlice)
}

// zero-out emptied values
//   - newSlice is at beginning of array with the new length
//   - slicedAway is the previously slicedAway slice
//   - elements to zero are at at end of slicedAway
func zeroOut[T any](newSlice, slicedAway []T) {

	// offset his how many elements slicedAway is off from newSlice
	var offset, isValid = Offset(newSlice, slicedAway)
	if !isValid {
		return // some issue with slices
	}

	// number of elements to zero out at end of slicedAway
	var newEnd = len(newSlice)
	var oldEnd = offset + len(slicedAway)
	if newEnd >= oldEnd {
		return // no elements to zero-out
	}
	var elementCount = oldEnd - newEnd

	var zeroValue T
	for index := len(slicedAway) - elementCount; index < len(slicedAway); index++ {
		slicedAway[index] = zeroValue
	}
}

// Offset calculates how many items a slice-away slice is off
// from the initiaal slice
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
	var slice0p = *((*uint)(unsafe.Pointer(&slice0)))
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
	var offsetu = (slicedAwayp - slice0p) / elementSize
	if isValid =
		// byte difference must be even divisible by element size
		slice0p+offsetu*elementSize == slicedAwayp &&
			// slicedAway and offest cannot beyond the end of slice0
			offset+len(slicedAway) <= cap(slice0); //
	!isValid {
		return // bad offset or not divisible by element size
	}
	offset = int(offsetu)

	return
}
