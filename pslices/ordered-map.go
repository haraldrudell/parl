/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/constraints"
)

type OrderedMap[K constraints.Ordered, V any] struct {
	// orderedList provides order for iteration of values
	list parl.OrderedValues[K] // List()
	// m provides O(1) access to values
	m map[K]V
}

// NewOrderedMap is a key-value map ordered by key.
// NewOrderedMap cannot be used for key types slice map func.
// For key types slice map func or not constraints.Ordered, use NewOrderedMapAny.
// For custom sort order, use NewOrderedFunc.
// if V is large-sized struct, V can be pointer.
//  ordered := NewOrderedMap[string, string]()
func NewOrderedMap[K constraints.Ordered, V any]() (o1 parl.OrderedMap[K, V]) {
	return &OrderedMap[K, V]{
		list: NewOrderedValues[K](),
		m:    map[K]V{},
	}
}

// NewOrderedMapFunc is a key-value map ordered by key using a function.
// For key types slice map func or not constraints.Ordered, use NewOrderedMapAny.
// if V is large-sized struct, V can be pointer.
func NewOrderedMapFunc[K constraints.Ordered, V any](cmp func(a, b K) (result int)) (o1 parl.OrderedMap[K, V]) {
	if cmp == nil {
		var keyValue K
		if cmpObject, ok := any(keyValue).(Comparable[K]); ok {
			cmp = cmpObject.Cmp
		}
		if cmp == nil {
			panic(perrors.Errorf("NewOrdered with cmp nil K: %T", keyValue))
		}
	}
	return &OrderedMap[K, V]{
		list: NewOrderedAny(cmp),
		m:    map[K]V{},
	}
}

// Get retrives a value O(1).
// if New is non-nil, an element is created if not already existing.
// if New is nil, a non-existing key returns nil.
func (o1 *OrderedMap[K, V]) Get(key K,
	newV func() (value *V),
	makeV func() (value V)) (value V) {
	var ok bool
	if value, ok = o1.m[key]; ok {
		return // found key return
	}
	if newV == nil && makeV == nil {
		return // no key, no newV or makeV: nil return
	}
	if newV != nil {
		pt := newV()
		if pt == nil {
			panic(perrors.New("Ordered newV: returned nil"))
		}
		o1.m[key] = *pt
	} else {
		o1.m[key] = makeV()
	}
	o1.list.Insert(key)

	return
}

func (o1 OrderedMap[K, V]) Has(key K) (value V, ok bool) {
	value, ok = o1.m[key]
	return
}

func (o1 OrderedMap[K, V]) Delete(key K) {
	delete(o1.m, key)
	o1.list.Delete(key)
}

func (o1 *OrderedMap[K, V]) List() (list []V) {
	list = make([]V, len(o1.m))
	for i, key := range o1.list.List() {
		list[i] = o1.m[key]
	}
	return
}
