/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ints

import (
	"errors"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
	"golang.org/x/exp/constraints"
)

// can be tested: errors.Is(err, ints.ErrToLarge)
var ErrTooLarge = errors.New("value too large")

// can be tested: errors.Is(err, ints.ErrNegative)
var ErrNegative = errors.New("negative value")

// Unsigned converts integer of type T to an unsigned value that fits the size of unsigned integer U
//   - U is the unsigned integer returned
//   - T is the input integer
//   - integer T is of any integer size and may be signed
//   - unsigned U is any size unsigned integer
//   - error: if integer is negative or too large to be held in type U
func Unsigned[U constraints.Unsigned, T constraints.Integer](integer T, label string) (unsigned U, err error) {

	// check for negative value
	//	- those come in signed integers
	if isSigned, _, _, _ := IntProperties[T](); isSigned {
		// whatever it is, it will fit into int64
		if i64 := int64(integer); i64 < 0 {
			if label == "" {
				label = pruntime.PackFunc()
			}
			err = perrors.Errorf("%s %w: %d -0x%x", label, ErrNegative, i64, -i64)
			return
		}
	}
	// integer is a positive number of a signed or unsigned integer

	// check that positive value fits
	var u64 = uint64(integer)
	_, max, _, _ := IntProperties[U]()
	if u64 > max {
		if label == "" {
			label = pruntime.PackFunc()
		}
		err = perrors.Errorf("%s %w: %d 0x%[1]x max: %d 0x%[2]x", label, ErrTooLarge, u64, max)
		return
	}

	unsigned = U(u64)

	return
}
