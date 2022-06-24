/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"sync"

	"github.com/haraldrudell/parl"
	"golang.org/x/exp/constraints"
)

type OrderedThreadSafe[K constraints.Ordered, V any] struct {
	lock    sync.Mutex
	ordered parl.OrderedMap[K, V]
}

func NewOrderedThreadSafe[K constraints.Ordered, V constraints.Ordered]() (o1 parl.OrderedMap[K, V]) {
	return &OrderedThreadSafe[K, V]{
		ordered: NewOrderedMap[K, V](),
	}
}

func NewOrderedFuncThreadSafe[K constraints.Ordered, V any](cmp func(a, b K) (result int)) (o1 parl.OrderedMap[K, V]) {
	return &OrderedThreadSafe[K, V]{
		ordered: NewOrderedMapFunc[K, V](cmp),
	}
}

func (o1 *OrderedThreadSafe[K, V]) Get(
	key K,
	newV func() (value *V),
	makeV func() (value V)) (value V) {
	o1.lock.Lock()
	defer o1.lock.Unlock()

	return o1.ordered.Get(key, newV, makeV)
}

func (o1 *OrderedThreadSafe[K, V]) Has(key K) (value V, ok bool) {
	o1.lock.Lock()
	defer o1.lock.Unlock()

	return o1.ordered.Has(key)
}

func (o1 *OrderedThreadSafe[K, V]) Delete(key K) {
	o1.lock.Lock()
	defer o1.lock.Unlock()

	o1.ordered.Delete(key)
}

func (o1 *OrderedThreadSafe[K, V]) List() (list []V) {
	o1.lock.Lock()
	defer o1.lock.Unlock()

	return o1.ordered.List()
}
