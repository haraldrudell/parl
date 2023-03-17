/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package ints provide manipulation of integer types.
package ints

import (
	"errors"

	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/constraints"
)

var ErrTooLarge = errors.New("value too large")
var ErrNegative = errors.New("negative value")

// Unsigned converts integer of type T to an unsigned value that fits the size of unsigned integer U
//   - integer T is of any integer size and may be signed
//   - unsigned U is any size unsigned integer
//   - error: if integer is negative or too large to be held in type U
func Unsigned[T constraints.Integer, U constraints.Unsigned](integer T, label string) (unsigned U, err error) {

	// check for negative value
	if isSigned, _, _, _ := IntProperties[T](); isSigned {
		if i64 := int64(integer); i64 < 0 {
			if label == "" {
				label = perrors.PackFunc()
			}
			err = perrors.Errorf("%s %w: %d -0x%x", label, ErrNegative, i64, -i64)
			return
		}
	}

	// check that positive value fits
	u64 := uint64(integer)
	_, max, _, _ := IntProperties[U]()
	if u64 > max {
		if label == "" {
			label = perrors.PackFunc()
		}
		err = perrors.Errorf("%s %w: %d 0x%[1]x max: %d 0x%[2]x", label, ErrTooLarge, u64, max)
		return
	}

	unsigned = U(u64)

	return
}
