/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import (
	"strconv"
	"testing"

	"github.com/haraldrudell/parl/iters"
)

func TestFuncSet(t *testing.T) {
	var value1, value2 = basics(1), basics(2)
	var value1d = strconv.Itoa(int(value1))
	var values = []basics{value1, value2}
	var eIDFunc = func(ep *basics) (t basics) { return *ep }

	var isValid, hasValue bool
	var value, zeroValue basics
	var iterator iters.Iterator[basics]
	var full string

	// IsValid() Iterator() Description() StringT() String()
	var set SetID[basics, basics]
	var reset = func() {
		// E is basics
		// T is int
		set = NewFunctionSet(values, eIDFunc)
	}

	// IsValid of element should return true
	reset()
	isValid = set.IsValid(value1)
	if !isValid {
		t.Error("IsValid false")
	}

	// IsValid of non-element should return true
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
	if full != value1d {
		t.Errorf("Description %q exp %q", full, value1d)
	}

	// StringT
	reset()
	var _ = (&FuncSet[int, string]{}).StringT
	full = set.StringT(value1)
	if full != value1d {
		t.Errorf("StringT %q exp %q", full, value1d)
	}
}
