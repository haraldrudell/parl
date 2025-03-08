/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"testing"
)

func TestAtomicMax(t *testing.T) {
	var threshold1, value1, value2 = 1, 1, 2

	var value, zeroValue int
	var hasValue, isNewMax, isPanic bool
	var err error

	var a AtomicMax[int]
	_ = 1

	// a new should have no value
	a = AtomicMax[int]{}
	value, hasValue = a.Max()
	if value != zeroValue {
		t.Errorf("new value %d exp %d", value, zeroValue)
	}
	if hasValue {
		t.Error("new hasValue true")
	}

	// Value below threshold should not be a max
	NewAtomicMaxp(&a, threshold1)
	isNewMax = a.Value(zeroValue)
	if isNewMax {
		t.Error("below isNewMax")
	}
	value, hasValue = a.Max()
	if value != zeroValue {
		t.Errorf("below value %d exp %d", value, zeroValue)
	}
	if hasValue {
		t.Error("below hasValue true")
	}

	// zero should be max
	a = AtomicMax[int]{}
	isNewMax = a.Value(zeroValue)
	if !isNewMax {
		t.Error("zero isNewMax false")
	}
	value, hasValue = a.Max()
	if value != zeroValue {
		t.Errorf("zero value %d exp %d", value, zeroValue)
	}
	if !hasValue {
		t.Error("zero hasValue false")
	}

	// zero-zero should not be max
	a = AtomicMax[int]{}
	isNewMax = a.Value(zeroValue)
	_ = isNewMax
	isNewMax = a.Value(zeroValue)
	if isNewMax {
		t.Error("zero isNewMax false")
	}

	// equal to threshold should be max
	NewAtomicMaxp(&a, threshold1)
	isNewMax = a.Value(value1)
	if !isNewMax {
		t.Error("equal isNewMax false")
	}
	value, hasValue = a.Max()
	if value != value1 {
		t.Errorf("equal value %d exp %d", value, value1)
	}
	if !hasValue {
		t.Error("equal hasValue false")
	}

	// smaller value should not be max
	a = AtomicMax[int]{}
	isNewMax = a.Value(value2)
	_ = isNewMax
	isNewMax = a.Value(value1)
	if isNewMax {
		t.Error("smaller isNewMax")
	}
	value, hasValue = a.Max()
	if value != value2 {
		t.Errorf("smaller value %d exp %d", value, value2)
	}
	if !hasValue {
		t.Error("smaller hasValue false")
	}

	// max1 should work
	a = AtomicMax[int]{}
	isNewMax = a.Value(value1)
	_ = isNewMax
	value = a.Max1()
	if value != value1 {
		t.Errorf("Max1 value %d exp %d", value, value1)
	}

	// negative threshold should panic
	isPanic, err = invokeNewAtomicMax()
	if !isPanic {
		t.Error("negative threshold no panic")
	}
	if err == nil {
		t.Error("negative threshold no error")
	}

	// negative value should panic
	isPanic, err = invokeAtomicMaxValue()
	if !isPanic {
		t.Error("negative value no panic")
	}
	if err == nil {
		t.Error("negative value no error")
	}
}

func invokeNewAtomicMax() (isPanic bool, err error) {
	defer PanicToErr(&err, &isPanic)

	NewAtomicMax(-1)

	return
}

func invokeAtomicMaxValue() (isPanic bool, err error) {
	defer PanicToErr(&err, &isPanic)

	NewAtomicMax(0).Value(-1)

	return
}
