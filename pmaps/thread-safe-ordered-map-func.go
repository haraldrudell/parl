/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// ThreadSafeOrderedMapFunc is a mapping whose values are provided in custom order. Thread-safe.
package pmaps

import (
	"github.com/haraldrudell/parl/parli"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
)

// ThreadSafeOrderedMapFunc is a mapping whose values are provided in custom order. Thread-safe.
type ThreadSafeOrderedMapFunc[K comparable, V any] struct {
	ThreadSafeMap[K, V]
	list parli.Ordered[V]
	cmp  func(a, b V) (result int)
}

func NewThreadSafeOrderedMapFunc[K comparable, V any](
	cmp func(a, b V) (result int),
) (orderedMap *ThreadSafeOrderedMapFunc[K, V]) {
	return &ThreadSafeOrderedMapFunc[K, V]{
		ThreadSafeMap: *NewThreadSafeMap[K, V](),
		list:          pslices.NewOrderedAny(cmp),
		cmp:           cmp,
	}
}

// NewThreadSafeOrderedMapFunc2 returns a mapping whose values are provided in custom order.
func NewThreadSafeOrderedMapFunc2[K comparable, V constraints.Ordered](
	list parli.Ordered[V],
) (orderedMap *ThreadSafeOrderedMapFunc[K, V]) {
	if list == nil {
		panic(perrors.NewPF("list cannot be nil"))
	} else if list.Length() > 0 {
		list.Clear()
	}
	return &ThreadSafeOrderedMapFunc[K, V]{
		ThreadSafeMap: *NewThreadSafeMap[K, V](),
		list:          list,
		cmp:           compare[V],
	}
}

func compare[V constraints.Ordered](a, b V) (result int) {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}
	return 0
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
func (m *ThreadSafeOrderedMapFunc[K, V]) GetOrCreate(
	key K,
	newV func() (value *V),
	makeV func() (value V),
) (value V, ok bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	mp := m.ThreadSafeMap.m

	// try existing mapping
	if value, ok = mp[key]; ok {
		return // mapping exists return
	}

	// create using newV
	if newV != nil {
		pt := newV()
		if pt == nil {
			panic(perrors.NewPF("newV returned nil"))
		}
		value = *pt
		mp[key] = value
		ok = true
		return // created using newV return
	}

	// create using makeV
	if makeV != nil {
		value = makeV()
		mp[key] = value
		ok = true
		return // created using makeV return
	}

	return // no key, no newV or makeV: nil return
}

// Put saves or replaces a mapping
func (m *ThreadSafeOrderedMapFunc[K, V]) Put(key K, value V) {
	m.lock.Lock()
	defer m.lock.Unlock()

	length0 := len(m.ThreadSafeMap.m)
	m.ThreadSafeMap.m[key] = value
	if length0 == len(m.ThreadSafeMap.m) {
		return
	}
	m.list.Insert(value)
}

// Put saves or replaces a mapping
//   - if mapping exists and poutif i non-nil, puIf function is invoked
//   - put is only carried out if mapping is new or putIf is non-nil and returns true
func (m *ThreadSafeOrderedMapFunc[K, V]) PutIf(key K, value V, putIf func(value V) (doPut bool)) (wasNewKey bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	value0, ok := m.ThreadSafeMap.m[key]
	if wasNewKey = !ok; !wasNewKey {
		if putIf == nil || !putIf(value0) {
			return // existing key, putIf nil or returning false: do nothing
		}
		if m.cmp(value0, value) != 0 {
			m.list.Delete(value0)
			m.list.Insert(value)
		}
	}
	m.ThreadSafeMap.m[key] = value

	return
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (m *ThreadSafeOrderedMapFunc[K, V]) Delete(key K) {
	m.lock.Lock()
	defer m.lock.Unlock()

	var v V
	var ok bool
	if v, ok = m.ThreadSafeMap.m[key]; !ok {
		return
	}
	delete(m.ThreadSafeMap.m, key)
	m.list.Delete(v)
}

// Clear empties the map
func (m *ThreadSafeOrderedMapFunc[K, V]) Clear() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.ThreadSafeMap.m = make(map[K]V)
	m.list.Clear()
}

// Clone returns a shallow clone of the map
func (m *ThreadSafeOrderedMapFunc[K, V]) Clone() (clone *ThreadSafeOrderedMapFunc[K, V]) {
	clone = NewThreadSafeOrderedMapFunc[K, V](m.cmp)
	m.lock.RLock()
	defer m.lock.RUnlock()

	clone.ThreadSafeMap.m = maps.Clone(m.ThreadSafeMap.m)
	clone.list = m.list.Clone()
	return
}

// List provides the mapped values in order
//   - O(n)
func (m *ThreadSafeOrderedMapFunc[K, V]) List(n ...int) (list []V) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.list.List(n...)
}
