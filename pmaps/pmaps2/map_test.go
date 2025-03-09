/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps2

import (
	"testing"
)

type testMapValue struct{ value int }

func TestMap(t *testing.T) {
	const (
		expLength = 2
	)
	var (
		v1        = testMapValue{1}
		v2        = testMapValue{2}
		v3        = testMapValue{3}
		zeroValue *testMapValue
	)

	var (
		m     Map[int, *testMapValue]
		m2    *Map[int, *testMapValue]
		value *testMapValue
		ok    bool
	)

	NewMap(&m)
	m.Put(v1.value, &v1)
	m.Put(v2.value, &v2)
	m.Put(v1.value, &v1)

	// Length should return number of elements
	if m.Length() != expLength {
		t.Errorf("Length %d exp %d", m.Length(), expLength)
	}

	// Get should return the corresponding mapping
	value, ok = m.Get(v2.value)
	if !ok {
		t.Error("ok false")
	}
	if value != &v2 {
		t.Errorf("Get %v exp %v", value, &v2)
	}

	// Get for non-existing mapping is zero-value, false
	value, ok = m.Get(v3.value)
	if ok {
		t.Error("ok true")
	}
	if value != zeroValue {
		t.Errorf("Get2 %v exp %v", value, zeroValue)
	}

	// Clone should return duplicate
	m2 = m.Clone2()
	if m2.Length() != expLength {
		t.Errorf("Length m2 %d exp %d", m2.Length(), expLength)
	}
}
