/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package ints provide manipulation of integer types.
package ints

import (
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

func TestUnsigned(t *testing.T) {
	var value = uint64(2)

	var u64 uint64
	var err error

	if u64, err = Unsigned[uint64](value, ""); err != nil {
		t.Errorf("err: %s", perrors.Short(err))
	}
	if u64 != value {
		t.Errorf("actual: %d exp %d", u64, value)
	}
}
