/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"testing"

	"github.com/haraldrudell/parl/parli"
	"golang.org/x/exp/slices"
)

func TestNewOrderedPointers(t *testing.T) {
	v1 := 1
	v1p := &v1
	v2 := 2
	v2p := &v2
	exp := []*int{v1p, v1p, v2p}

	var orderedPointers parli.OrderedPointers[int]
	var actual []*int
	cmp := func(a, b *int) (result int) {
		if *a < *b {
			return -1
		} else if *a > *b {
			return 1
		}
		return 0
	}

	orderedPointers = NewOrderedPointers[int]()

	orderedPointers.Insert(v2p)
	orderedPointers.Insert(v1p)
	orderedPointers.Insert(v1p)
	actual = orderedPointers.List()
	if slices.CompareFunc(actual, exp, cmp) != 0 {
		t.Errorf("bad slice: %v exp %v", ResolveSlice(orderedPointers.List()), ResolveSlice(exp))
	}
}
