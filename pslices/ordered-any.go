/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// OrderedAny is a slice ordered by a function allowing duplicates.
// OrderedAny implements [parl.Ordered][E any].
package pslices

import (
	"github.com/haraldrudell/parl"
	"golang.org/x/exp/slices"
)

// OrderedAny is a slice ordered by a function allowing duplicates.
// OrderedAny implements [parl.Ordered][E any].
//   - cmp allows for custom ordering or ordering of slice map and function types
//   - Use E as value for small-sized data or interface values.
//   - Use E as a pointer for larger sized structs.
//   - Duplicates are allowed, Inssert places duplicates at end
//   - Insert and Delete O(log n)
//   - cmp(a, b) is expected to return an integer comparing the two parameters:
//     0 if a == b, a negative number if a < b and a positive number if a > b
type OrderedAny[E any] struct {
	Slice[E] // Element() Length() List() Clear()
	// cmp is a comparison function returning <0 if a<b, 0 if a == b and 1 otherwise.
	// if cmp is nil, order is based on comparison of values
	cmp func(a, b E) (result int)
}

// NewOrderedAny creates a list ordered by a comparison function.
//   - The cmp comparison function can be provided by E being pslices.Comparable,
//     ie. a type having a Cmp method.
//   - cmp(a, b) is expected to return an integer comparing the two parameters:
//     0 if a == b, a negative number if a < b and a positive number if a > b
//   - duplicate values are allowed and inserted in order with later values last
func NewOrderedAny[E any](cmp func(a, b E) (result int)) (list parl.Ordered[E]) {
	if cmp == nil {
		var e E
		cmp = CmpFromComparable(e)
	}

	return &OrderedAny[E]{cmp: cmp}
}

// Insert adds a value to the ordered slice.
func (o *OrderedAny[E]) Insert(element E) {
	o.list = InsertOrderedFunc(o.list, element, o.cmp)
}

// Delete removes an element from the ordered slice.
//   - if the element did not exist, the slice is not changed
//   - if element exists in duplicates, a random element of those duplicates is removed
//   - O(log n)
func (o *OrderedAny[E]) Delete(element E) {
	if position, wasFound := slices.BinarySearchFunc(o.list, element, o.cmp); wasFound {
		o.list = slices.Delete(o.list, position, position+1)
	}
}

func (o *OrderedAny[E]) Index(element E) (index int) {
	var wasFound bool
	if index, wasFound = slices.BinarySearchFunc(o.list, element, o.cmp); !wasFound {
		index = -1
	}
	return
}

// Length returns the number of elements
func (o *OrderedAny[E]) Clone() (o2 parl.Ordered[E]) {
	return &OrderedAny[E]{
		Slice: Slice[E]{list: slices.Clone(o.list)},
		cmp:   o.cmp,
	}
}
