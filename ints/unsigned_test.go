/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ints

import (
	"errors"
	"math"
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

func TestUnsigned(t *testing.T) {
	var u8max, i8min, i8max = uint8(math.MaxUint8), int8(math.MinInt8), int8(math.MaxInt8)
	var u64one, u64max = uint64(1), uint64(math.MaxUint64)

	var u64 uint64
	var u8, u8zeroValue uint8
	var err error

	// uint8 should fit uint64
	u64, err = Unsigned[uint64](u8max, "")
	if err != nil {
		t.Errorf("uint8 err: %s", perrors.Short(err))
	}
	if u64 != uint64(u8max) {
		t.Errorf("uint8 %d exp %d", u64, uint64(u8max))
	}

	// small uint64 should fit uint8
	u8, err = Unsigned[uint8](u64one, "")
	if err != nil {
		t.Errorf("small err: %s", perrors.Short(err))
	}
	if u8 != uint8(u64one) {
		t.Errorf("small %d exp %d", u8, uint8(u64one))
	}

	// positive int8 should fit uint8
	u8, err = Unsigned[uint8](i8max, "")
	if err != nil {
		t.Errorf("i8max err: %s", perrors.Short(err))
	}
	if u8 != uint8(i8max) {
		t.Errorf("i8max %d exp %d", u8, uint8(i8max))
	}

	// negative value should error
	u8, err = Unsigned[uint8](i8min, "")
	if err == nil {
		t.Error("negative missing error")
	} else if !errors.Is(err, ErrNegative) {
		t.Errorf("negative error not ErrNegative: %s", perrors.Short(err))
	}
	if u8 != u8zeroValue {
		t.Errorf("negative %d exp %d", u8, u8zeroValue)
	}

	// too large value should error
	u8, err = Unsigned[uint8](u64max, "")
	if err == nil {
		t.Error("large missing error")
	} else if !errors.Is(err, ErrTooLarge) {
		t.Errorf("large error not ErrTooLarge: %s", perrors.Short(err))
	}
	if u8 != u8zeroValue {
		t.Errorf("large %d exp %d", u8, u8zeroValue)
	}
}
