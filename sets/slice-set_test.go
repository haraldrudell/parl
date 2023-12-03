/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import (
	"testing"

	"github.com/haraldrudell/parl/iters"
)

func TestSliceSet(t *testing.T) {
	type e struct{ s string }
	var value1, value2 = 0, 2
	var name1, name2 = "n1", "n2"

	var value1full string
	var value1stringT = "{" + name1 + "}"
	var values = []e{{name1}, {name2}}

	var isValid, hasValue bool
	var value int
	var full string
	var iterator iters.Iterator[int]
	var elementType *e
	var eIterator iters.Iterator[*e]

	// IsValid() Iterator() Description() StringT() String()
	var set SetID[int, e]
	var reset = func() {
		set = NewSliceSet(values)
	}

	// IsValid of element should return true
	reset()
	isValid = set.IsValid(value1)
	if !isValid {
		t.Error("IsValid false")
	}

	// IsValid of non-element should return false
	reset()
	isValid = set.IsValid(value2)
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

	// Element should return element
	reset()
	elementType = set.Element(value1)
	if elementType != &values[0] {
		t.Error("set.Element bad")
	}

	// EIterator should iterate
	reset()
	eIterator = set.EIterator()
	elementType, hasValue = eIterator.Next()
	_ = hasValue
	if elementType != &values[0] {
		t.Error("EIterator.Next bad")
	}
}
