/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"testing"

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
