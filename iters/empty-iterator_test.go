/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import (
	"testing"
)

func TestEmptyIterator(t *testing.T) {
	var ok bool
	var zeroValue int
	var actual int
	var err error

	iter := NewEmptyIterator[int]()

	if actual, ok = iter.Next(); ok {
		t.Error("Next returned true")
	} else if actual != zeroValue {
		t.Error("Next returned other than zero-value")
	}
	if ok = iter.HasNext(); ok {
		t.Error("GoNext returned true")
	}
	if actual = iter.NextValue(); actual != zeroValue {
		t.Errorf("Next not zero value: %d exp %d", actual, zeroValue)
	}
	if actual, ok = iter.Same(); ok {
		t.Error("Next returned true")
	} else if actual != zeroValue {
		t.Error("Next returned other than zero-value")
	}
	if ok = iter.Has(); ok {
		t.Error("Has returned true")
	}
	if actual = iter.SameValue(); actual != zeroValue {
		t.Errorf("Same not zero value: %d exp %d", actual, zeroValue)
	}
	if err = iter.Cancel(); err != nil {
		t.Errorf("Cancel err: %v", err)
	}
}
