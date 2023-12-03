/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package iters

import (
	"testing"
)

func TestEmpty(t *testing.T) {
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
	if actual, ok = iter.Same(); ok {
		t.Error("Next returned true")
	} else if actual != zeroValue {
		t.Error("Next returned other than zero-value")
	}
	if err = iter.Cancel(); err != nil {
		t.Errorf("Cancel err: %v", err)
	}
}
