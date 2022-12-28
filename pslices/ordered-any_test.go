/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// OrderedAny is a slice ordered by a function alllowing duplicates.
package pslices

import (
	"testing"

	"github.com/haraldrudell/parl"
	"golang.org/x/exp/slices"
)

func TestNewOrderedAny(t *testing.T) {
	v1 := 1
	v2 := 2
	exp2 := []int{v1, v2}

	var list parl.Ordered[int]
	var list2 parl.Ordered[int]
	var actual []int
	var actInt int
	cmp := func(a, b int) (result int) {
		if a > b {
			return 1
		} else if a < b {
			return -1
		}
		return 0
	}

	list = NewOrderedAny(cmp)

	list.Insert(v2)
	list.Insert(v1)
	actual = list.List()
	if slices.Compare(actual, exp2) != 0 {
		t.Errorf("bad slice %v exp %v", actual, exp2)
	}

	list.Delete(v1)
	if actInt = list.Index(v2); actInt != 0 {
		t.Errorf("bad Index %d exp %d", actInt, 0)
	}
	if actInt = list.Index(v1); actInt != -1 {
		t.Errorf("bad Index2 %d exp %d", actInt, -1)
	}

	list2 = list.Clone()
	if slices.Compare(list.List(), list2.List()) != 0 {
		t.Errorf("bad Clone %v exp %v", list.List(), list2.List())
	}

}
