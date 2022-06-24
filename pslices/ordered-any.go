/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/slices"
)

// OrderedAny is an ordered list of values sorted using a function.
// Use E as value for small-sized data or interface values.
// Use E as a pointer for large-sized structs.
// Insert overwrites duplicates.
type OrderedAny[E any] struct {
	list []E
	// cmp is a comparison function returning <0 if a<b, 0 if a == b and 1 otherwise.
	// if cmp is nil, order is based on comparison of values
	cmp func(a, b E) (result int)
}

// NewOrderedAny creates a list ordered by a comparison function.
// cmp can be provided by E being pslices.Comparable.
func NewOrderedAny[E any](cmp func(a, b E) (result int)) (list parl.OrderedValues[E]) {
	if cmp == nil {
		var element E
		if cmpObject, ok := any(element).(Comparable[E]); ok {
			cmp = cmpObject.Cmp
		}
		if cmp == nil {
			panic(perrors.New("NewOrderedAny with cmp nil"))
		}
	}
	return &OrderedAny[E]{cmp: cmp}
}

func (o *OrderedAny[E]) Insert(element E) {
	o.list = InsertOrderedFunc(o.list, element, o.cmp)
}

func (o *OrderedAny[E]) Delete(element E) {
	if position, wasFound := slices.BinarySearchFunc(o.list, element, o.cmp); wasFound {
		o.list = slices.Delete(o.list, position, position+1)
	}
}

func (o *OrderedAny[E]) List() (list []E) {
	return slices.Clone(o.list)
}
