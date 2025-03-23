/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import "unsafe"

// Offset calculates how many elements slicedAway is sliced off
// from the start of slice0, the initial slice
//   - slice0 is a slice of non-zero capacity at
//     the beginning of the underlying array
//   - slicedAway is a slice of non-zero capacity that
//     is from a slice-expression of slice0
//   - offset is how any elements slicedAway is sliced off of the start o slice0
//   - isValid true: offset is valid
//   - isValid false: the operation failed:
//   - — the capacity of slice0 or slicedAway is zero
//   - — slice0 and slicedAway are not backed by the same array,
//     ie. they are disparate slices
func Offset[T any](slice0, slicedAway []T) (offset int, isValid bool) {
	var capSlice0 = cap(slice0)

	// verify that operation is possible
	//	- if slice capacity is zero, the slice’s array pointer is undefined
	//	- therefore, both slice0 and slicedAway must have non-zero capacity
	if capSlice0 == 0 || cap(slicedAway) == 0 {
		return // operation failed: isValid false
	}
	// both slice0 and slicedAway have valid array pointers

	// slice0First is pointer-value to the first value of the underlying array
	//	- because capacity is non-zero, the array is guaranteed to exist
	//	- uint type to do pointer arithmetic
	var slice0First = *(*uint)(unsafe.Pointer(&slice0))
	// slicedAwayFirst is pointer-value to the first value of slicedAway
	var slicedAwayFirst = *(*uint)(unsafe.Pointer(&slicedAway))

	// an array of two T elements
	var s [2]T
	// elementByteSize is the number of bytes separating each element
	//	- small positive integer
	var elementByteSize = //
	int(uintptr(unsafe.Pointer(&s[1]))) -
		int(uintptr(unsafe.Pointer(&s[0])))

		// byteOffset is how many bytes off slicedAwayFirst is from slice0First
	var byteOffset = int(slicedAwayFirst) - int(slice0First)
	var offset0 = byteOffset / elementByteSize
	if isValid = byteOffset >= 0 &&
		byteOffset%elementByteSize == 0 &&
		offset0 < capSlice0; !isValid {
		return // corrupt offset return: isValid false
	}

	offset = offset0
	return // success: offset valid, isValid true
}
