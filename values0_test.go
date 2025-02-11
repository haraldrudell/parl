/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"testing"
)

func TestNewValues(t *testing.T) {
	const (
		value1, value2                  = 1, 2
		expCount0, expCount1, expCount2 = 0, 1, 2
	)
	var (
		count      int
		valueSlice = []int{value1, value2}
	)
	// Add() Count() Seq()
	var values Values[int] = NewValues[int]()

	// count should be zero
	count = values.Count()
	if count != expCount0 {
		t.Errorf("Count %d exp %d", count, expCount0)
	}

	// count after Add should be 1
	values.Add(value1)
	count = values.Count()
	if count != expCount1 {
		t.Errorf("Count %d exp %d", count, expCount1)
	}

	// count after 2Add should be 2
	values.Add(value2)
	count = values.Count()
	if count != expCount2 {
		t.Errorf("Count %d exp %d", count, expCount2)
	}

	// Seq should iterate over the added values
	count = 0
	for v := range values.Seq {
		count++
		if count > len(valueSlice) {
			t.Fatalf("too many iterations: %d exp %d", count, len(valueSlice))
		}
		if v != valueSlice[count-1] {
			t.Errorf("bad Seq index %d: %d exp %d", count-1, v, valueSlice[count-1])
		}
	}
	if count != len(valueSlice) {
		t.Errorf("Too few iterations: %d exp %d", count, len(valueSlice))
	}
}
