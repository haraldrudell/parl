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

func TestBasicSet(t *testing.T) {
	var value1, value2 = basics(1), basics(2)
	var value1d = strconv.Itoa(int(value1))
	var values = []basics{value1, value2}

	var isValid, hasValue bool
	var value, zeroValue basics
	var iterator iters.Iterator[basics]
	var full string

	// IsValid() Iterator() Description() StringT() String()
	var set Set[basics]
	var reset = func() {
		set = NewBasicSet(values)
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
	if full != value1d {
		t.Errorf("Description %q exp %q", full, value1d)
	}

	// StringT
	reset()
	full = set.StringT(value1)
	if full != value1d {
		t.Errorf("StringT %q exp %q", full, value1d)
	}
}

type basics int

func (i basics) Description() (s string) { return strconv.Itoa(int(i)) }
