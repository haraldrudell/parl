/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package ints provide manipulation of integer types.
package ints

import (
	"math"
	"strconv"
	"testing"
	"unsafe"

	"golang.org/x/exp/constraints"
)

func TestIntProperties(t *testing.T) {

	// uint8
	var isSignedU8 = false
	var maxU8 = uint64(math.MaxUint8)
	var maxNegativeU8 = int64(0)
	var sizeofU8 = unsafe.Sizeof(uint8(0))

	// int64
	var isSignedI64 = true
	var maxI64 = uint64(math.MaxInt64)
	var maxNegativeI64 = int64(math.MinInt64)
	var sizeofI64 = unsafe.Sizeof(int64(0))

	// uint64
	var isSignedU64 = false
	var maxU64 = uint64(math.MaxUint64)
	var maxNegativeU64 int64
	var sizeofU64 = unsafe.Sizeof(uint64(0))

	var isSigned bool
	var maxPositive uint64
	var maxNegative int64
	var sizeof int

	var u8 uint8
	isSigned, maxPositive, maxNegative, sizeof = IntProperties(u8)
	if isSigned != isSignedU8 {
		t.Error("isSignedU8")
	}
	if maxPositive != maxU8 {
		t.Error("maxU8")
	}
	if maxNegative != maxNegativeU8 {
		t.Error("maxNegativeU8")
	}
	if sizeof != int(sizeofU8) {
		t.Error("sizeofU8")
	}

	isSigned, maxPositive, maxNegative, sizeof = IntProperties[int64]()
	if isSigned != isSignedI64 {
		t.Error("isSignedI64")
	}
	if maxPositive != maxI64 {
		t.Errorf("maxPositive: %s", signedHexadecimal(maxPositive))
	}
	if maxNegative != maxNegativeI64 {
		t.Errorf("maxNegativeI64: %s", signedHexadecimal(maxNegative))
	}
	if sizeof != int(sizeofI64) {
		t.Error("sizeofI64")
	}

	isSigned, maxPositive, maxNegative, sizeof = IntProperties[uint64]()
	if isSigned != isSignedU64 {
		t.Error("isSignedU64")
	}
	if maxPositive != maxU64 {
		t.Errorf("maxPositive: %s", signedHexadecimal(maxU64))
	}
	if maxNegative != maxNegativeU64 {
		t.Errorf("maxNegativeU64: %s", signedHexadecimal(maxNegative))
	}
	if sizeof != int(sizeofU64) {
		t.Error("sizeofU64")
	}

}

// signedHexadecimal returns a human-readable hexadecimal string
//   - positive → "0xffff_ffff_ffff_f800"
//   - negative → ""
func signedHexadecimal[T constraints.Integer](integer T) (hexString string) {
	var u64 uint64
	var isNegative bool
	if isNegative = integer < 0; isNegative {
		u64 = uint64(-integer)
	} else {
		u64 = uint64(integer)
	}
	hexString = "0x" + strconv.FormatUint(u64, 16)
	if isNegative {
		hexString = "-" + hexString
	}
	for i := len(hexString) - 4; i > 2; i -= 4 {
		hexString = hexString[:i] + "_" + hexString[i:]
	}
	return
}
