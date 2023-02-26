/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// ThreadSafeOrderedMapFunc is a mapping whose values are provided in custom order. Thread-safe.
package pmaps

import (
	"sync"

	"github.com/haraldrudell/parl/perrors"
)

// ThreadSafeOrderedMapFunc is a mapping whose values are provided in custom order. Thread-safe.
type ThreadSafeOrderedMapFunc[K comparable, V any] struct {
	lock sync.RWMutex
	OrderedMapFunc[K, V]
}

func NewThreadSafeOrderedMapFunc[K comparable, V any](
	cmp func(a, b V) (result int),
) (orderedMap *ThreadSafeOrderedMapFunc[K, V]) {
	return &ThreadSafeOrderedMapFunc[K, V]{
		OrderedMapFunc: *NewOrderedMapFunc[K](cmp),
	}
}

func (mp *ThreadSafeOrderedMapFunc[K, V]) Get(key K) (value V, ok bool) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.OrderedMapFunc.Get(key)
}

// GetOrCreate returns an item from the map if it exists otherwise creates it.
//   - newV or makeV are invoked in the critical section, ie. these functions
//     may not access the map or deadlock
//   - if a key is mapped, its value is returned
//   - otherwise, if newV and makeV are both nil, nil is returned.
//   - otherwise, if newV is present, it is invoked to return a pointer ot a value.
//     A nil return value from newV causes panic. A new mapping is created using
//     the value pointed to by the newV return value.
//   - otherwise, a mapping is created using whatever makeV returns
//   - newV and makeV may not access the map.
//     The map’s write lock is held during their execution
//   - GetOrCreate is an atomic, thread-safe operation
//   - value insert is O(log n)
func (rw *ThreadSafeOrderedMapFunc[K, V]) GetOrCreate(
	key K,
	newV func() (value *V),
	makeV func() (value V),
) (value V, ok bool) {
	rw.lock.Lock()
	defer rw.lock.Unlock()

	// try existing mapping
	if value, ok = rw.OrderedMapFunc.Get(key); ok {
		return // mapping exists return
	}

	// create using newV
	if newV != nil {
		pt := newV()
		if pt == nil {
			panic(perrors.NewPF("newV returned nil"))
		}
		value = *pt
		rw.OrderedMapFunc.Put(key, value)
		ok = true
		return // created using newV return
	}

	// create using makeV
	if makeV != nil {
		value = makeV()
		rw.OrderedMapFunc.Put(key, value)
		ok = true
		return // created using makeV return
	}

	return // no key, no newV or makeV: nil return
}

// Put saves or replaces a mapping
func (mp *ThreadSafeOrderedMapFunc[K, V]) Put(key K, value V) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.OrderedMapFunc.Put(key, value)
}

// Put saves or replaces a mapping
func (mp *ThreadSafeOrderedMapFunc[K, V]) PutIf(key K, value V, putIf func(value V) (doPut bool)) (wasNewKey bool) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	existing, keyExists := mp.OrderedMapFunc.Get(key)
	wasNewKey = !keyExists
	if keyExists && putIf != nil && !putIf(existing) {
		return // putIf false return: this value should not be updated
	}
	mp.OrderedMapFunc.Put(key, value)

	return
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (mp *ThreadSafeOrderedMapFunc[K, V]) Delete(key K) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.OrderedMapFunc.Delete(key)
}

// Clear empties the map
func (mp *ThreadSafeOrderedMapFunc[K, V]) Clear() {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.OrderedMapFunc.Clear()
}

func (mp *ThreadSafeOrderedMapFunc[K, V]) Length() (length int) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.OrderedMapFunc.Length()
}

// Clone returns a shallow clone of the map
func (mp *ThreadSafeOrderedMapFunc[K, V]) Clone() (clone *ThreadSafeOrderedMapFunc[K, V]) {

	// there is no access to cmp, so create an empty clone that has a lock
	//	- any write to clone must happen while its lock is held
	c := ThreadSafeOrderedMapFunc[K, V]{}
	clone = &c

	// get write lock for clone
	c.lock.Lock()
	defer c.lock.Unlock()

	// get read lock for mp
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	// clone map and list, including cmp
	c.OrderedMapFunc = *mp.OrderedMapFunc.Clone()

	return
}

// List provides the mapped values in order
//   - O(n)
func (mp *ThreadSafeOrderedMapFunc[K, V]) List(n ...int) (list []V) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.OrderedMapFunc.List(n...)
}
