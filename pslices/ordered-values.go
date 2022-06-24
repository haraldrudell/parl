/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"github.com/haraldrudell/parl"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

// OrderedValues is an ordered list of values sorted by value.
// E must be a comparable type, ie. not slice func map.
// Insert overwrites duplicates.
// for custom sort order, use NewOrderedAny
type OrderedValues[E constraints.Ordered] struct {
	list []E
}

func NewOrderedValues[E constraints.Ordered]() (list parl.OrderedValues[E]) {
	return &OrderedValues[E]{}
}

func (o *OrderedValues[E]) Insert(element E) {
	if position, wasFound := slices.BinarySearch(o.list, element); wasFound {
		o.list[position] = element
	} else {
		o.list = slices.Insert(o.list, position, element)
	}
}

func (o *OrderedValues[E]) Delete(element E) {
	if position, wasFound := slices.BinarySearch(o.list, element); wasFound {
		o.list = slices.Delete(o.list, position, position+1)
	}
}

func (o *OrderedValues[E]) List() (list []E) {
	return slices.Clone(o.list)
}
