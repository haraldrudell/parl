/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"github.com/haraldrudell/parl/parli"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pmaps/pmaps2"
)

// threadSafeMap is a private promotable field
// that does not promote any nonn-interface public identifiers
//   - native Go map functions: Get Put Delete Length Range
//   - convenience methods: clone Clear
//   - order methods: List
type threadSafeMap[K comparable, V any] struct {
	tsm pmaps2.ThreadSafeMap[K, V]
}

// newThreadSafeMap returns a thread-safe Go map
func newThreadSafeMap[K comparable, V any](fieldp *threadSafeMap[K, V]) (m *threadSafeMap[K, V]) {

	// set m
	if m = fieldp; m == nil {
		m = &threadSafeMap[K, V]{}
	}

	// initialize all fields
	pmaps2.NewThreadSafeMap(&m.tsm)

	return
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
func (m *threadSafeMap[K, V]) GetOrCreate(
	key K,
	newV func() (value *V),
	makeV func() (value V),
) (value V, hasValue bool) {
	defer m.tsm.Lock().Unlock()

	// try existing mapping
	if value, hasValue = m.tsm.Get(key); hasValue {
		return // mapping exists return
	}

	// create using newV
	if newV != nil {
		pt := newV()
		if pt == nil {
			panic(perrors.NewPF("newV returned nil"))
		}
		value = *pt
		m.tsm.Put(key, value)
		hasValue = true
		return // created using newV return
	}

	// create using makeV
	if makeV != nil {
		value = makeV()
		m.tsm.Put(key, value)
		hasValue = true
		return // created using makeV return
	}

	return // no key, no newV or makeV: nil return
}

// clone is package-private helper method providing access to encapsulated Clone method
func (m *threadSafeMap[K, V]) clone(tsm *threadSafeMap[K, V]) { m.tsm.Clone(&tsm.tsm) }

// cloneToGomap is package-private helper providing access to
// encapsulated thread-safe map
func (m *threadSafeMap[K, V]) cloneToGomap(goMap *map[K]V) { m.tsm.CloneToGoMap(goMap) }

func (m *threadSafeMap[K, V]) Get(key K) (value V, hasValue bool) {
	defer m.tsm.RLock().RUnlock()

	return m.tsm.Get(key)
}

func (m *threadSafeMap[K, V]) Put(key K, value V) {
	defer m.tsm.Lock().Unlock()

	m.tsm.Put(key, value)
}

func (m *threadSafeMap[K, V]) Delete(key K, useZeroValue ...parli.DeleteMethod) {
	defer m.tsm.Lock().Unlock()

	m.tsm.Delete(key, useZeroValue...)
}

func (m *threadSafeMap[K, V]) Length() (length int) {
	defer m.tsm.RLock().RUnlock()

	return m.tsm.Length()
}

func (m *threadSafeMap[K, V]) Range(rangeFunc func(key K, value V) (keepGoing bool)) (rangedAll bool) {
	defer m.tsm.RLock().RUnlock()

	return m.tsm.Range(rangeFunc)
}

// Clear empties the map
func (m *threadSafeMap[K, V]) Clear(useRange ...parli.ClearMethod) {
	defer m.tsm.Lock().Unlock()

	m.tsm.Clear(useRange...)
}

// List provides the mapped values, undefined ordering
//   - O(n)
func (m *threadSafeMap[K, V]) List(n ...int) (list []V) {

	// get n
	var n0 int
	if len(n) > 0 {
		n0 = n[0]
	}
	defer m.tsm.RLock().RUnlock()

	return m.tsm.List(n0)
}
