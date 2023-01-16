/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package ints provide manipulation of integer types.
package ints

import (
	"errors"
	"math"

	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/constraints"
)

var ErrTooLarge = errors.New("value too large")

func ConvertU8[T constraints.Integer](integer T, label string) (u8 uint8, err error) {

	// convert to uint64
	var u64 uint64
	if u64, err = ConvertU64(integer, label); err != nil {
		return
	}

	// convert to uint8
	if u64 > math.MaxUint8 {
		if label == "" {
			label = perrors.PackFunc()
		}
		err = perrors.ErrorfPF("%s %w: %d 0x%[1]x max: %d 0x%[2]x", label, ErrTooLarge, u64, math.MaxUint8)
		return
	}

	u8 = uint8(u64)

	return
}

func ConvertU16[T constraints.Integer](integer T, label string) (u16 uint16, err error) {

	// convert to uint64
	var u64 uint64
	if u64, err = ConvertU64(integer, label); err != nil {
		return
	}

	// convert uint64 to uint16
	if u64 > math.MaxUint16 {
		if label == "" {
			label = perrors.PackFunc()
		}
		err = perrors.Errorf("%s value too large: %d 0x%[1]x max: %d 0x%[2]x", label, u64, math.MaxUint16)
		return
	}

	u16 = uint16(u64)

	return
}

func ConvertU32[T constraints.Integer](integer T, label string) (u32 uint32, err error) {

	// convert to uint64
	var u64 uint64
	if u64, err = ConvertU64(integer, label); err != nil {
		return
	}

	// convert uint64 to uint16
	if u64 > math.MaxUint32 {
		if label == "" {
			label = perrors.PackFunc()
		}
		err = perrors.Errorf("%s value too large: %d 0x%[1]x max: %d 0x%[2]x", label, u64, math.MaxUint32)
		return
	}

	u32 = uint32(u64)

	return
}

func ConvertU64[T constraints.Integer](integer T, label string) (u64 uint64, err error) {
	var T0 T
	var u64Max uint64 = math.MaxUint64
	var Tmax = T(u64Max)
	isSigned := Tmax < T0
	if isSigned {
		i64 := int64(integer)
		if i64 < 0 {
			if label == "" {
				label = perrors.PackFunc()
			}
			err = perrors.Errorf("%s negative value: %d -0x%x", label, i64, -i64)
			return
		}
		u64 = uint64(i64)
		return
	}

	u64 = uint64(integer)

	return
}
