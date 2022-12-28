/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"testing"

	"github.com/haraldrudell/parl"
	"golang.org/x/exp/slices"
)

func TestInsertOrdered(t *testing.T) {
	v1 := 1
	exp := []int{v1}
	exp2 := []int{v1, v1}

	var slice0 []int
	var slice1 []int
	var slice2 []int

	slice1 = InsertOrdered(slice0, v1)
	if slices.Compare(slice1, exp) != 0 {
		t.Errorf("bad slice1 %v exp %v", slice1, exp)
	}

	slice2 = InsertOrdered(slice1, v1)
	if slices.Compare(slice2, exp2) != 0 {
		t.Errorf("bad slice2 %v exp %v", slice1, exp)
	}

	//	{"insert", args{[]int{1, 3}, 2}, []int{1, 2, 3}},
}

func TestInsertOrderedFunc(t *testing.T) {
	v1 := 1
	v2 := 2
	exp1 := []int{v1}
	exp2 := []int{v2, v1}
	exp3 := []int{v2, v1, v1}

	var slice0 []int
	var slice1 []int
	var slice2 []int
	var slice3 []int
	descending := func(a, b int) (result int) {
		if a < b {
			return 1
		} else if a > b {
			return -1
		}
		return 0
	}

	if slice1 = InsertOrderedFunc(slice0, v1, descending); slices.Compare(slice1, exp1) != 0 {
		t.Errorf("bad slice1 %v exp %v", slice1, exp1)
	}
	if slice2 = InsertOrderedFunc(slice1, v2, descending); slices.Compare(slice2, exp2) != 0 {
		t.Errorf("bad slice2 %v exp %v", slice2, exp2)
	}
	if slice3 = InsertOrderedFunc(slice2, v1, descending); slices.Compare(slice3, exp3) != 0 {
		t.Errorf("bad slice3 %v exp %v", slice3, exp3)
	}
}

type CmpPointer struct {
	value int
	ID    int
}

var _ parl.Comparable[CmpPointer] = &CmpPointer{}

func (e *CmpPointer) Cmp(b CmpPointer) (result int) {
	if e.value > b.value {
		return 1
	} else if e.value < b.value {
		return -1
	}
	return 0
}

func TestInsertOrderedFunc_CmpPointer(t *testing.T) {
	e1 := CmpPointer{1, 1}
	var sliceA = []CmpPointer{e1}
	expB := []CmpPointer{e1, e1}

	// slice of struct: comparable but not Constraints.Ordered
	var sliceB []CmpPointer
	_ = sliceB

	sliceB = InsertOrderedFunc(sliceA, e1, nil)
	// slices.Compare requires Constraint.Ordered
	different := len(sliceB) != len(expB)
	if !different {
		for i, e := range sliceB {
			if different = e != expB[i]; different {
				break
			}
		}
	}
	if different {
		t.Errorf("bad sliceB %v exp %v", sliceB, expB)
	}
}

type CmpValue struct {
	value int
	ID    int
}

var _ parl.Comparable[CmpValue] = CmpValue{}

func (e CmpValue) Cmp(b CmpValue) (result int) {
	if e.value > b.value {
		return 1
	} else if e.value < b.value {
		return -1
	}
	return 0
}

func TestInsertOrderedFunc_CmpValue(t *testing.T) {
	e1 := CmpValue{1, 1}
	var sliceA = []CmpValue{e1}
	expB := []CmpValue{e1, e1}

	var sliceB []CmpValue
	_ = sliceB

	sliceB = InsertOrderedFunc(sliceA, e1, nil)
	different := len(sliceB) != len(expB)
	if !different {
		for i, e := range sliceB {
			if different = e != expB[i]; different {
				break
			}
		}
	}
	if different {
		t.Errorf("bad sliceB %v exp %v", sliceB, expB)
	}
}
