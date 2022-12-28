/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// RWMap is a thread-safe mapping.
package pmaps

import (
	"testing"

	"github.com/haraldrudell/parl"
	"golang.org/x/exp/slices"
)

func TestNewRWMap(t *testing.T) {
	k1 := "key1"
	v1 := 1
	k2 := "key2"
	v2 := 2
	exp := []int{v1, v2}
	exp2 := []int{v2}

	var m parl.ThreadSafeMap[string, int]
	var m2 parl.ThreadSafeMap[string, int]
	var list []int
	var list2 []int
	newV := func() (value *int) { return &v1 }
	makeV := func() (value int) { return v2 }

	m = NewRWMap[string, int]()

	m.GetOrCreate(k1, newV, nil)
	m.GetOrCreate(k1, nil, nil)
	m.GetOrCreate(k2, nil, makeV)
	list = m.List()
	slices.Sort(list)
	if slices.Compare(list, exp) != 0 {
		t.Errorf("bad list %v exp %v", list, exp)
	}

	m.Delete(k1)
	m.Put(k2, v2)
	list = m.List()
	slices.Sort(list)
	if slices.Compare(list, exp2) != 0 {
		t.Errorf("bad list2 %v exp %v", list, exp2)
	}

	m2 = m.Clone()
	list2 = m2.List()
	slices.Sort(list2)
	if slices.Compare(list, list2) != 0 {
		t.Errorf("bad clone %v exp %v", list2, list)
	}

	m.Clear()
	if m.Length() != 0 {
		t.Errorf("bad length %d exp %d", m.Length(), 0)

	}

	(&RWMap[string, int]{}).GetOrCreate(k1, nil, nil)
	(&RWMap[string, int]{}).Get(k1)
	(&RWMap[string, int]{}).List()
	(&RWMap[string, int]{}).Put(k1, v1)
	(&RWMap[string, int]{}).Delete(k1)
	(&RWMap[string, int]{}).Clear()
	(&RWMap[string, int]{}).Length()
}
