/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"github.com/haraldrudell/parl/parli"
	"golang.org/x/exp/constraints"
)

// OrderedPointers is an ordered list of *E pointers sorted by the referenced values.
// OrderedPointers implements [parl.OrderedPointers][E any].
//   - The OrderedPointers ordered list does not require a comparison function
//   - E is used for large structs.
//   - Insert overwrites duplicates.
//   - for custom sort order, use NewOrderedAny
type OrderedPointers[E constraints.Ordered] struct {
	OrderedAny[*E] // Element() Length() List() Clear() Insert() Delete() Index() Clone()
}

func NewOrderedPointers[E constraints.Ordered]() (list parli.OrderedPointers[E]) {
	var o = OrderedPointers[E]{}
	o.OrderedAny = *NewOrderedAny(o.Cmp).(*OrderedAny[*E])
	return &o
}

func (o *OrderedPointers[E]) Cmp(a, b *E) (result int) {
	if *a < *b {
		return -1
	} else if *a > *b {
		return 1
	}
	return 0
}
