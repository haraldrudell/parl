/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import (
	"testing"

	"github.com/haraldrudell/parl/iters"
)

func TestSet(t *testing.T) {
	var value1, value2 = 1, 2
	var name1, name2 = "n1", "n2"

	var value1full string
	var value1stringT = name1
	var values = []SetElement[int]{
		{ValueV: value1, Name: name1},
		{ValueV: value2, Name: name2},
	}

	var isValid, hasValue bool
	var value, zeroValue int
	var iterator iters.Iterator[int]
	var full string

	// IsValid() Iterator() Description() StringT() String()
	var set Set[int]
	var reset = func() {
		set = NewSet[int](values)
	}

	// IsValid of element should return true
	reset()
	isValid = set.IsValid(value1)
	if !isValid {
		t.Error("IsValid false")
	}

	// IsValid of non-element should return false
	reset()
	isValid = set.IsValid(zeroValue)
	if isValid {
		t.Error("IsValid true")
	}

	// Iterator should iterate
	reset()
	iterator = set.Iterator()
	value, hasValue = iterator.Next()
	_ = hasValue
	if value != value1 {
		t.Errorf("Iterator.Next %d exp %d", value, value1)
	}

	// Description
	reset()
	full = set.Description(value1)
	if full != value1full {
		t.Errorf("Description %q exp %q", full, value1full)
	}

	// StringT
	reset()
	full = set.StringT(value1)
	if full != value1stringT {
		t.Errorf("StringT %q exp %q", full, value1stringT)
	}
}
