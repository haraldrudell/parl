/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import (
	"testing"
)

func TestConverterIterator(t *testing.T) {
	// keys that the converter-iterator will iterate over
	var keys = []string{"z", "keyTwo"}
	// the expected values produces by the converter iterator
	var values = func(keys []string) (values []int) {
		values = make([]int, len(keys))
		for i, key := range keys {
			values[i] = len(key)
		}
		return
	}(keys)

	t.Logf("keys: %v", keys)
	t.Logf("values: %v", values)

	var keyIterator = NewSliceIterator(keys)
	var value int
	var hasValue bool
	var zeroValue int
	var err error

	var iterator Iterator[int] = NewConverterIterator(
		keyIterator,
		converterFunction,
	)

	// Same should advance to the first value
	value, hasValue = iterator.Same()
	//hasValue should be true
	if !hasValue {
		t.Error("Same hasValue false")
	}
	// value should be first value
	if value != values[0] {
		t.Errorf("Same value %q exp %q", value, values[0])
	}

	// Next should return the second value
	value, hasValue = iterator.Next()
	if !hasValue {
		t.Errorf("Next hasValue false")
	}
	if value != values[1] {
		t.Errorf("Next value %q exp %q", value, values[1])
	}

	// Next should return no value
	value, hasValue = iterator.Next()
	if hasValue {
		t.Errorf("Next2 hasValue true")
	}
	if value != zeroValue {
		t.Errorf("Next2 value %q exp %q", value, zeroValue)
	}

	// cancel should not return error
	if err = iterator.Cancel(); err != nil {
		t.Errorf("Cancel err '%v'", err)
	}
}

// type ConverterFunction[K constraints.Ordered, V any]
// func(key K, isCancel bool) (value V, err error)
var _ ConverterFunction[string, int] = converterFunction

// converterFunction can be used with a
// ConverterIterator as ConverterFunction[string, int]
func converterFunction(key string, isCancel bool) (value int, err error) {

	// handle cancel
	if isCancel {
		return
	}

	// produce value corresponding to key
	value = len(key)

	return
}
