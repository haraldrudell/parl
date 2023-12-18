/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"slices"
	"testing"
	"unsafe"
)

func TestShifter(t *testing.T) {
	// the slice for all subsequent slices
	var slice0 = []int{1, 2}
	// index to use for slicing away at the beginning
	var sliceAwayIndex = 1
	// items to append
	var items = []int{3}
	// expected resulting slice
	var sliceExp = []int{2, 3}

	// Append()
	var shifter *Shifter[int] = NewShifter(slice0)

	shifter.Slice = shifter.Slice[sliceAwayIndex:]
	shifter.Append(items...)

	// resulting slice should match
	if !slices.Equal(shifter.Slice, sliceExp) {
		t.Errorf("Append %v exp %v", shifter.Slice, sliceExp)
	}

	// slice should not have been re-allocated
	var shifterp = int(uintptr(unsafe.Pointer(&shifter.Slice[0])))
	var slicep = int(uintptr(unsafe.Pointer(&slice0[0])))
	if shifterp != slicep {
		t.Errorf("shifterp\n0x%x exp\n0x%x", shifterp, slicep)
	}
}

func TestShifterZeroFill(t *testing.T) {
	// the initial slice for all subsequent slices
	var slice0 = []int{1, 2, 3}
	// index to use for slice away at the beginning
	var sliceAway = 2
	// items to append
	var items = []int{4}
	// expected slice result
	var sliceExp = []int{3, 4}
	// index in slice0 where zero-fill should have taken place
	var zeroFillIndex = 2

	var zeroValue int

	// Append()
	var shifter *Shifter[int] = NewShifter(slice0, ZeroFillingShifter)

	// slice away at beginning
	shifter.Slice = shifter.Slice[sliceAway:]
	// append at end
	shifter.Append(items...)

	// slice result should match
	if !slices.Equal(shifter.Slice, sliceExp) {
		t.Errorf("Append %v exp %v", shifter.Slice, sliceExp)
	}

	// slice should not have been reallocated
	var shifterp = int(uintptr(unsafe.Pointer(&shifter.Slice[0])))
	var slicep = int(uintptr(unsafe.Pointer(&slice0[0])))
	if shifterp != slicep {
		t.Errorf("shifterp\n0x%x exp\n0x%x", shifterp, slicep)
	}

	// slice element should have been zero-filled
	if slice0[zeroFillIndex] != zeroValue {
		t.Error("zero-fill failed")
	}
}
