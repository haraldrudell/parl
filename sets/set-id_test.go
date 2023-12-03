/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import (
	"testing"

	"github.com/haraldrudell/parl/iters"
)

func TestSetID(t *testing.T) {
	var value1, value2 = 1, 2
	var name1, name2 = "n1", "n2"

	var values = []SetElement[int]{
		{ValueV: value1, Name: name1},
		{ValueV: value2, Name: name2},
	}

	var hasValue bool
	var value, elementType *SetElement[int]
	var eIterator iters.Iterator[*SetElement[int]]

	// IsValid() Iterator() Description() StringT() String()
	var set SetID[int, SetElement[int]]
	var reset = func() {
		set = NewSetID[int](values)
	}

	// Element of element should return true
	reset()
	elementType = set.Element(value1)
	if elementType != &values[0] {
		t.Error("set.Element bad")
	}

	// Iterator should iterate
	reset()
	eIterator = set.EIterator()
	value, hasValue = eIterator.Next()
	_ = hasValue
	if value != &values[0] {
		t.Error("Iterator.Next bad")
	}
}
