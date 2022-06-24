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

type OrderedMapAny[O any, K constraints.Ordered, V any] struct {
	// orderedList provides ordered access by type O and function cmp
	list parl.OrderedValues[O] // List()
	// m provides O(1) access to values
	m map[K]V
	// getKey converts O any to K that can be used as a generic map key
	getKey func(order O) (key K)
}

// NewOrderedMapAny is a key-value map for any type of key that is ordered by O any using a function.
// Order can be by O value or by custom order using cmp
// getKey converts the O any to a constraints.Ordered key value.
// if V is large-sized struct, V can be pointer.
func NewOrderedMapAny[O any, K constraints.Ordered, V any](
	cmp func(a, b O) (result int),
	getKey func(order O) (key K),
) (o1 parl.OrderedMapAny[O, K, V]) {

	// determine ordering method
	if cmp == nil {
		// try to get cmp from Comparable
		var orderValue O
		if cmpObject, ok := any(orderValue).(Comparable[O]); ok {
			cmp = cmpObject.Cmp
		}
		if cmp == nil {
			// if there is no cmp, O must be constraints.Ordered so NewOrderedMap should be used
			panic(perrors.Errorf("NewOrderedMapAny with cmp nil: use NewOrderedMap instead. O: %T", orderValue))
		}
	}
	if getKey == nil {
		panic(perrors.New("NewOrderedMapAny with getKey nil"))
	}
	return &OrderedMapAny[O, K, V]{
		list:   NewOrderedAny(cmp),
		m:      map[K]V{},
		getKey: getKey,
	}
}

func (o1 *OrderedMapAny[O, K, V]) Get(order O,
	newV func() (value *V),
	makeV func() (value V)) (value V) {
	key := o1.getKey(order)
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
		value = *pt
	} else {
		value = makeV()
	}
	o1.m[key] = value
	o1.list.Insert(order)
	return
}

func (o1 *OrderedMapAny[O, K, V]) Has(order O) (value V, ok bool) {
	value, ok = o1.m[o1.getKey(order)]
	return
}

func (o1 *OrderedMapAny[O, K, V]) Delete(order O) {
	delete(o1.m, o1.getKey(order))
	o1.list.Delete(order)
}

func (o1 *OrderedMapAny[O, K, V]) List() (list []V) {
	list = make([]V, len(o1.m))
	for i, oAny := range o1.list.List() {
		list[i] = o1.m[o1.getKey(oAny)]
	}
	return
}
