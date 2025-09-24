/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ints

import (
	"math"
	"strconv"
	"testing"
	"unsafe"

	"golang.org/x/exp/constraints"
)

func TestIntProperties(t *testing.T) {
	//t.Fail()

	const cIsUnsigned, cIsSigned = false, true
	const cMinIsZero = 0
	type IntPropertiesFunc func() (isSigned bool, maxPositive uint64, maxNegative int64, sizeof int)
	var typs = []struct {
		basicTypeName     string
		intPropertiesFunc IntPropertiesFunc
		isSigned          bool
		maxPositive       uint64
		maxNegative       int64
		sizeof            int
	}{
		{"uint8", func() (bool, uint64, int64, int) { return IntProperties(uint8(0)) }, cIsUnsigned, math.MaxUint8, cMinIsZero, int(unsafe.Sizeof(uint8(0)))},
		{"int8", func() (bool, uint64, int64, int) { return IntProperties(int8(0)) }, cIsSigned, math.MaxInt8, math.MinInt8, int(unsafe.Sizeof(int8(0)))},
		{"uint64", func() (bool, uint64, int64, int) { return IntProperties(uint64(0)) }, cIsUnsigned, math.MaxUint64, cMinIsZero, int(unsafe.Sizeof(uint64(0)))},
		{"int64", func() (bool, uint64, int64, int) { return IntProperties(int64(0)) }, cIsSigned, math.MaxInt64, math.MinInt64, int(unsafe.Sizeof(int64(0)))},
		{"uintptr", func() (bool, uint64, int64, int) { return IntProperties(uintptr(0)) }, cIsUnsigned, math.MaxUint64, cMinIsZero, int(unsafe.Sizeof(uintptr(0)))},
	}
	for _, typ := range typs {
		var isSigned, maxPositive, maxNegative, sizeof = typ.intPropertiesFunc()
		t.Logf("%s maxPositive: %s maxNegative: %s", typ.basicTypeName, signedHexadecimal(maxPositive), signedHexadecimal(maxNegative))
		if isSigned != typ.isSigned {
			t.Errorf("%s isSigned %t", typ.basicTypeName, isSigned)
		}
		if maxPositive != typ.maxPositive {
			t.Errorf("%s maxPositive %d exp %d", typ.basicTypeName, maxPositive, typ.maxPositive)
		}
		if maxNegative != typ.maxNegative {
			t.Errorf("%s maxNegative %d exp %d", typ.basicTypeName, maxNegative, typ.maxNegative)
		}
		if sizeof != typ.sizeof {
			t.Errorf("%s sizeof %d exp %d", typ.basicTypeName, sizeof, typ.sizeof)
		}
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
