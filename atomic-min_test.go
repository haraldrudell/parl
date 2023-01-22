/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"math"
	"testing"
)

func TestAtomicMin(t *testing.T) {
	var value1 uint64 = math.MaxUint64 - 2
	var value2 uint64 = 1
	var value3 uint64 = 3

	var value uint64
	var hasValue bool

	var min AtomicMin[uint64]

	if _, hasValue = min.Min(); hasValue {
		t.Error("1 hasValue true")
	}

	if !min.Value(value1) {
		t.Error("2 not min")
	}

	if value, hasValue = min.Min(); !hasValue {
		t.Error("3 hasValue false")
	}
	if value != value1 {
		t.Errorf("3 value %d exp %d", value, value1)
	}

	if !min.Value(value2) {
		t.Error("4 not min")
	}

	if value, hasValue = min.Min(); !hasValue {
		t.Error("5 hasValue false")
	}
	if value != value2 {
		t.Errorf("5 value %d exp %d", value, value2)
	}

	if min.Value(value3) {
		t.Error("6 min")
	}

	if value, hasValue = min.Min(); !hasValue {
		t.Error("7 hasValue false")
	}
	if value != value2 {
		t.Errorf("7 value %d exp %d", value, value2)
	}

}
