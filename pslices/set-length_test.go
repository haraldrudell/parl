/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"slices"
	"testing"

	"github.com/haraldrudell/parl/pslices/pslib"
)

func TestSetLength(t *testing.T) {
	var slice1, expSlice0, slice2 = []int{1}, []int{0}, []int{0, 0}
	var length1, length2 = 1, 2

	var slice, slice0 []int

	// extend length from nil
	slice = nil
	SetLength(&slice, 1)
	if !slices.Equal(slice, expSlice0) {
		t.Errorf("SetLength 1 for nil: %v exp%v", slice, expSlice0)
	}

	// SetLength noop should do nothing to do return
	slice = nil
	SetLength(&slice, 0)
	if slice != nil {
		t.Errorf("SetLength 0 for nil non-nil: %v", slice)
	}

	// noZero should work
	slice0 = slices.Clone(slice1)
	slice = slice0
	SetLength(&slice, 0, pslib.NoZeroOut)
	if !slices.Equal(slice0, slice1) {
		t.Errorf("SetLength 1 for nil: %v exp%v", slice, slice1)
	}

	// shortening with doZero
	slice0 = slices.Clone(slice1)
	slice = slice0
	// SetLength 0 should zero-out element
	SetLength(&slice, 0)
	if !slices.Equal(slice0, expSlice0) {
		t.Errorf("SetLength 0 for 1-slice: %v exp%v", slice, expSlice0)
	}

	// extend to cap
	slice0 = slices.Clone(expSlice0)
	slice = slice0[:0]
	SetLength(&slice, length2)
	if !slices.Equal(slice, slice2) {
		t.Errorf("SetLength 2 for 1-cap slice: %v exp%v", slice, slice2)
	}

	// extend within cap
	slice0 = slices.Clone(slice1)
	slice = slice0[:0]
	SetLength(&slice, length1)
	if !slices.Equal(slice, expSlice0) {
		t.Errorf("SetLength within cap: %v exp%v", slice, expSlice0)
	}
}
