/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import (
	"slices"
	"testing"
)

// p converts a string pointer to a string value
func p(value *string) (s string) {
	if value != nil {
		return *value
	}
	return "nil"
}

func TestNewSlicePointerIterator(t *testing.T) {
	var values = []string{"one", "two"}

	var err error
	var value *string
	var hasValue bool
	var zeroValue *string

	var iterator Iterator[*string] = NewSlicePointerIterator(slices.Clone(values))

	// request IsSame value twice should:
	//	- retrieve the first value and return it
	//	- then return the same value again
	for i := 0; i <= 1; i++ {
		value, hasValue = iterator.Same()

		//hasValue should be true
		if !hasValue {
			t.Errorf("Same%d hasValue false", i)
		}
		// value should be first value
		if value == nil || *value != values[0] {
			t.Errorf("Same%d value %q exp %q", i, p(value), values[0])
		}
	}

	// Next should return the second value
	value, hasValue = iterator.Next()
	if !hasValue {
		t.Errorf("Next hasValue false")
	}
	if value == nil || *value != values[1] {
		t.Errorf("Next value %q exp %q", p(value), values[1])
	}

	// Next should return no value
	value, hasValue = iterator.Next()
	if hasValue {
		t.Errorf("Next2 hasValue true")
	}
	if value != zeroValue {
		t.Errorf("Next2 value %q exp %q", p(value), p(zeroValue))
	}

	// cancel should not return error
	if err = iterator.Cancel(); err != nil {
		t.Errorf("Cancel err '%v'", err)
	}
}
