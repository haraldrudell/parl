/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ints

import (
	"math"

	"github.com/haraldrudell/parl/perrors"
)

var (
	// max value for unsigned half-int
	MaxUnsignedHalf int
	// max value for signed half-int
	MaxSignedHalf int
	// min value for signed half-int
	MinSignedHalf = func() (minSignedHalf int) {
		if is64bit {
			MaxUnsignedHalf = math.MaxUint32
			MaxSignedHalf = math.MaxInt32
			minSignedHalf = math.MinInt32
			return
		}
		MaxUnsignedHalf = math.MaxUint16
		MaxSignedHalf = math.MaxInt16
		minSignedHalf = math.MinInt16
		return
	}()
	maxUint64 = uint64(math.MaxUint64)
)

// Make combines two unsigned half-T ints into T
//   - littleHalf, bigHalf:
//   - — for unsigned halves, T 64-bit: limited to 0–math.MaxUint32
//   - — for unsigned halves, T 32-bit limited to 0–math.MaxUint16
//   - — for signed halves, T 64-bit limited to math.MinInt32–math.MaxInt32
//   - — for signed halves, T 32-bit limited to math.MinInt16–math.MaxInt16
//   - combined: opaque value containing both halves
//   - —
//   - retrieve with [GetSigneds] or [GetUnsigneds]
//   - validate with [IsValidSignedHalf] [IsValidUnsignedHalf]
//   - T is uint (uint32 or uint64) or uint64
func Make[T ~uint | ~uint64](littleHalf, bigHalf int) (combined T) {
	var t_Is64Bits = uint64(T(maxUint64)) == math.MaxUint64
	if t_Is64Bits {
		combined = T(makeUint64(uint32(littleHalf), uint32(bigHalf)))
		return
	}
	combined = T(makeUint32(uint16(littleHalf), uint16(bigHalf)))
	return
}

// makeUint64 makes uint64 from two uint32
func makeUint64(littleHalf, bigHalf uint32) (combined uint64) {
	combined = uint64(littleHalf) | uint64(bigHalf)<<bits32
	return
}

// makeUint64 makes uint64 from two uint32
func makeUint32(littleHalf, bigHalf uint16) (combined uint32) {
	combined = uint32(littleHalf) | uint32(bigHalf)<<bits16
	return
}

// IsValidSignedHalfInt verifies range
//   - signedHalfInt an int to be used with [Make]
//   - err: non-nil if invalid
//   - for signed halves, T 64-bit limited to math.MinInt32–math.MaxInt32
//   - for signed halves, T 32-bit limited to math.MinInt16–math.MaxInt16
//   - T is uint (uint32 or uint64) or uint64
func IsValidSignedHalf[T ~uint | ~uint64](signedHalfInt int) (err error) {
	var t_Is64Bits = uint64(T(maxUint64)) == math.MaxUint64
	if t_Is64Bits {
		if signedHalfInt < math.MinInt32 || signedHalfInt > math.MaxInt32 {
			err = perrors.ErrorfPF("signedHalfInt out of range: %d [%d–%d]",
				signedHalfInt, math.MinInt32, math.MaxInt32,
			)
		}
		return
	}
	if signedHalfInt < math.MinInt16 || signedHalfInt > math.MaxInt16 {
		err = perrors.ErrorfPF("signedHalfInt out of range: %d [%d–%d]",
			signedHalfInt, math.MinInt16, math.MaxInt16,
		)
	}
	return
}

// IsValidUnsignedHalfInt verifies range
//   - err: non-nil if invalid
func IsValidUnsignedHalfInt[T ~uint | ~uint64](unsignedHalfInt int) (err error) {
	var t_Is64Bits = uint64(T(maxUint64)) == math.MaxUint64
	if t_Is64Bits {
		if unsignedHalfInt < 0 || unsignedHalfInt > math.MaxUint32 {
			err = perrors.ErrorfPF("unsigned half-int out of range: %d [%d–%d]",
				unsignedHalfInt, 0, math.MaxUint32,
			)
		}
		return
	}
	if unsignedHalfInt < 0 || unsignedHalfInt > math.MaxUint16 {
		err = perrors.ErrorfPF("unsigned half-int out of range: %d [%d–%d]",
			unsignedHalfInt, 0, math.MaxUint16,
		)
	}
	return
}

// GetSigneds splits uint64 into two signed limited ints
//   - for 64-bit limited to math.MinInt32–math.MaxInt32
//   - for 32-bit limited to math.MinInt16–math.MaxInt16
func GetSigneds[T ~uint | ~uint64](combined T) (littleHalf, bigHalf int) {
	var t_Is64Bits = uint64(T(maxUint64)) == math.MaxUint64
	if t_Is64Bits {
		littleHalf = int(int32(combined & math.MaxUint32))
		bigHalf = int(int32(combined >> bits32))
		return
	}
	littleHalf = int(int16(combined & math.MaxUint16))
	bigHalf = int(int16(combined >> bits16))
	return
}

// GetUnsigneds splits uint64 into two positive ints
//   - for 64-bit limited to 0–math.MaxUint32
//   - for 32-bit limited to 0–math.MaxUint16
func GetUnsigneds[T ~uint | ~uint64](combined T) (littleHalf, bigHalf int) {
	var t_Is64Bits = uint64(T(maxUint64)) == math.MaxUint64
	if t_Is64Bits {
		littleHalf = int((combined & math.MaxUint32))
		bigHalf = int((combined >> bits32))
		return
	}
	littleHalf = int(combined & math.MaxUint16)
	bigHalf = int(combined >> bits16)
	return
}

// CombinedUint contains two half-ints
type CombinedUint uint

// CombinedUint64 containing two half-ints
type CombinedUint64 uint64

const (
	// true if 64-bit system
	is64bit = math.MaxUint == math.MaxUint64
	// half number of bits of a uint64
	bits32 = 32
	// half number of bits of a uint32
	bits16 = 16
)
