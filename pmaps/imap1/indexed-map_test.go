package imap1_test

import (
	"slices"
	"strconv"
	"testing"

	"github.com/haraldrudell/parl/pmaps/imap1"
)

func TestIndexedMapNil(t *testing.T) {
	const (
		expLength1 = 1
	)
	var (
		value1  = "one"
		stringp = &value1
	)

	var (
		old        *string
		hadMapping bool
		length     int
	)

	var m imap1.IndexedMap[*int, *string] = imap1.MakeIndexedMap[*int, *string]()

	// Put 1 creates mapping
	old, hadMapping = m.Put(nil)
	if hadMapping {
		t.Error("FAIL Delete hadMapping true")
	}
	if old != nil {
		t.Error("FAIL Delete old exp nil")
	}
	length = m.Length()
	if length != expLength1 {
		t.Errorf("FAIL Length %d exp %d", length, expLength1)
	}

	// Put 2 with comparison same should be noop
	//	- check in debugger
	old, hadMapping = m.Put(nil)
	if !hadMapping {
		t.Error("FAIL Delete hadMapping false")
	}
	if old != nil {
		t.Errorf("FAIL Delete old %q exp nil", *old)
	}
	length = m.Length()
	if length != expLength1 {
		t.Errorf("FAIL Length %d exp %d", length, expLength1)
	}

	// Put 3 with comparison different should update
	old, hadMapping = m.Put(nil, stringp)
	if !hadMapping {
		t.Error("FAIL Delete hadMapping false")
	}
	if old != nil {
		t.Errorf("FAIL Delete old %q exp nil", *old)
	}
	length = m.Length()
	if length != expLength1 {
		t.Errorf("FAIL Length %d exp %d", length, expLength1)
	}
}

func TestIndexedMap(t *testing.T) {
	const (
		key1, key2, key3, key4 = 1, 2, 3, 4
		value1, value2, value3 = "one", "two", "three"
		expLength0             = 0
		expLength2             = 2
		expZeroValue           = ""
		index1, index2         = 1, 2
	)
	var (
		expKeyStrings = []string{strconv.Itoa(key1), strconv.Itoa(key2)}
		expKeys       = []int{key1, key2}
	)

	var (
		length     int
		hasValue   bool
		value      string
		keyStrings []string
		keys       []int
		old        string
		hadMapping bool
	)

	// Get() GetByIndex() Put() Delete() Contains() Length()
	// KeyStrings() Keys()
	var m imap1.IndexedMap[int, string] = imap1.MakeIndexedMap[int, string]()

	// Length() should be zero
	length = m.Length()
	if length != expLength0 {
		t.Errorf("FAIL Length %d exp %d", length, expLength0)
	}

	// Length() should be two
	// Put()
	m.Put(key1, value1)
	m.Put(key2, value2)
	length = m.Length()
	if length != expLength2 {
		t.Errorf("FAIL Length %d exp %d", length, expLength2)
	}

	// Contains should be false
	hasValue = m.Contains(key3)
	if hasValue {
		t.Error("FAIL Contains true")
	}

	// Contains should be true
	hasValue = m.Contains(key2)
	if !hasValue {
		t.Error("FAIL Contains false")
	}

	// Get should have hasValue false
	value, hasValue = m.Get(key3)
	if hasValue {
		t.Error("FAIL Get hasValue true")
	}
	if value != expZeroValue {
		t.Errorf("FAIL Get value %q exp %q", value, expZeroValue)
	}

	// Get should have hasValue true
	value, hasValue = m.Get(key1)
	if !hasValue {
		t.Error("FAIL Get hasValue false")
	}
	if value != value1 {
		t.Errorf("FAIL Get value %q exp %q", value, value1)
	}

	// GetByIndex should have hasValue false
	value, hasValue = m.GetByIndex(index2)
	if hasValue {
		t.Error("FAIL GetByIndex hasValue true")
	}
	if value != expZeroValue {
		t.Errorf("FAIL GetByIndex value %q exp %q", value, expZeroValue)
	}

	// GetByIndex should have hasValue true
	value, hasValue = m.GetByIndex(index1)
	if !hasValue {
		t.Error("FAIL GetByIndex hasValue false")
	}
	if value != value2 {
		t.Errorf("FAIL GetByIndex value %q exp %q", value, value2)
	}

	// Keystring should return keys
	keyStrings = m.KeyStrings()
	if !slices.Equal(keyStrings, expKeyStrings) {
		t.Errorf("FAIL KeyStrings %v exp %v", keyStrings, expKeyStrings)
	}

	// Keys should return keys
	keys = m.Keys()
	if !slices.Equal(keys, expKeys) {
		t.Errorf("FAIL KeyStrings %v exp %v", keys, expKeys)
	}

	// Delete should delete
	old, hadMapping = m.Delete(key1)
	if !hadMapping {
		t.Error("FAIL Delete hadMapping false")
	}
	if old != value1 {
		t.Errorf("FAIL Delete old %q exp %q", old, value1)
	}
}
