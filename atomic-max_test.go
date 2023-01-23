/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "testing"

func TestAtomicMax(t *testing.T) {
	var value1 uint64 = 2
	var value2 uint64 = 1
	var value3 uint64 = 3

	var max AtomicMax[uint64]

	if !max.Value(value1) {
		t.Error("value1: not max")
	}
	if max.Value(value2) {
		t.Error("value2: max")
	}
	if !max.Value(value3) {
		t.Error("value3: not max")
	}
	v, _ := max.Max()
	if v != value3 {
		t.Errorf("max %d exp %d", v, value3)
	}
}
