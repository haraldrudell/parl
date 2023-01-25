/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import "testing"

func TestDelegator(t *testing.T) {
	slice := []int{5, 6, 7, 8}

	var actualT int

	iter := NewSliceIterator(slice)

	// methods

	if !iter.HasNext() {
		t.Error("HasNext false")
	}

	if actualT = iter.NextValue(); actualT != slice[1] {
		t.Errorf("NextValue: %d exp %d", actualT, slice[1])
	}

	if !iter.Has() {
		t.Error("Has false")
	}

	if actualT = iter.SameValue(); actualT != slice[1] {
		t.Errorf("SameValue: %d exp %d", actualT, slice[1])
	}
}
