/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package ints provide manipulation of integer types.
package ints

import (
	"math"
	"unsafe"

	"golang.org/x/exp/constraints"
)

const (
	BitsPerByte = 8
)

// IntProperties returns the properties of integer I
//   - I is int int8 int16 int32 int64 uint
//     uint8 uint16 uint32 uint64 uintptr
//   - isSigned is true if I is a signed integer type
//   - maxPositive is the largest positive number that can be assigned to I
//   - maxNegative is the largest negative number that can be assigned to I.
//     For unsigned types, maxNegative is 0
//   - sizeof is the number of bytes I occupies: 1, 2, 4 or 8 for 8, 16, 32, 64-bit integers
func IntProperties[I constraints.Integer](variable ...I) (isSigned bool, maxPositive uint64, maxNegative int64, sizeof int) {

	// determine if I is signed
	maxPositive = math.MaxUint64
	// i’s highest bit is 1, so if i is signed, that is a negative number -1
	var i = I(maxPositive)
	sizeof = int(unsafe.Sizeof(i)) // Sizeof returns uintptr
	if isSigned = i < 0; isSigned {
		// I is signed: max negative number is highest bit 1 all other bits zero
		maxNegative = int64(I(1) << (sizeof*BitsPerByte - 1))
		// max positive number is highest bit 0 all other bits 1
		maxPositive = uint64(^I(maxNegative))
	} else {
		// I is unsigned: maxNegative is 0
		maxPositive = uint64(i)
	}

	return
}
