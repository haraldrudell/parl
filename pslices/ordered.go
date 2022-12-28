/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Ordered is an ordered slice overwriting duplicates implementing [parl.Ordered][E any].
package pslices

import (
	"github.com/haraldrudell/parl"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

// Ordered is an ordered slice overwriting duplicates implementing [parl.Ordered][E any].
//   - E must be a comparable and ordered type, ie. not slice func map.
//   - Insert overwrites duplicates and is O(log n)
//   - Delete removes the first occurrence O(log n)
//   - For custom sort order or slice func map types, use NewOrderedAny
type Ordered[E constraints.Ordered] struct {
	Slice[E]
}

// NewOrdered returns a slice ordered by value.
func NewOrdered[E constraints.Ordered]() (list parl.Ordered[E]) {
	return &Ordered[E]{}
}

// Insert adds an element to an ordered slice.
//   - Insert overwrites duplicates and is O(log n)
func (o *Ordered[E]) Insert(element E) {
	if position, wasFound := slices.BinarySearch(o.list, element); wasFound {
		o.list[position] = element
	} else {
		o.list = slices.Insert(o.list, position, element)
	}
}

// Delete removes an element from an ordered slice.
//   - if the element is not in the slice, the slice is not changed
//   - is the slice has duplicates, the first occurrence is removed
//   - O(log n)
func (o *Ordered[E]) Delete(element E) {
	if position, wasFound := slices.BinarySearch(o.list, element); wasFound {
		o.list = slices.Delete(o.list, position, position+1)
	}
}

func (o *Ordered[E]) Index(element E) (index int) {
	var wasFound bool
	if index, wasFound = slices.BinarySearch(o.list, element); !wasFound {
		index = -1
	}
	return
}

func (o *Ordered[E]) Clone() (o2 parl.Ordered[E]) {
	return &Ordered[E]{Slice: Slice[E]{list: slices.Clone(o.list)}}
}
