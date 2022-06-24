/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

// OrderedPointers is an ordered list of pointers sorted by the referenced values.
// E is used for large structs.
// Insert overwrites duplicates.
// for custom sort order, use NewOrderedAny
type OrderedPointers[E constraints.Ordered] struct {
	list []*E
}

func NewOrderedPointers[E constraints.Ordered]() (list parl.OrderedPointers[E]) {
	return &OrderedPointers[E]{}
}

func (o *OrderedPointers[E]) Insert(element *E) {
	if element == nil {
		panic(perrors.New("OrderedValue.Insert with element nil"))
	}
	if position, wasFound := slices.BinarySearchFunc(o.list, element, o.Cmp); wasFound {
		o.list[position] = element
	} else {
		o.list = slices.Insert(o.list, position, element)
	}
}

func (o *OrderedPointers[E]) List() (list []*E) {
	return slices.Clone(o.list)
}

func (o *OrderedPointers[E]) Cmp(a, b *E) (result int) {
	if *a < *b {
		return -1
	} else if *a > *b {
		return 1
	}
	return 0
}
