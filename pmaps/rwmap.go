/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// RWMap is a one-liner thread-safe mapping.
// RWMap implements [parli.ThreadSafeMap][K comparable, V any].
package pmaps

import (
	"github.com/haraldrudell/parl/parli"
)

// RWMap is a one-liner thread-safe mapping.
// RWMap implements [parli.ThreadSafeMap][K comparable, V any].
//   - GetOrCreate method is an atomic, thread-safe operation
//     as opposed to Get-then-Put
//   - PutIf is atomic, thread-safe operation
//   - native Go map functions: Get Put Delete Length Range
//   - convenience methods: Clone Clone2 Clear
//   - order functions: List Keys
//   - V is copied so if size of V is large or V contains locks, use pointer
//   - RWMap uses reader/writer mutual exclusion lock for slightly higher performance.
//   - Get methods are O(1)
type RWMap[K comparable, V any] struct {
	threadSafeMap[K, V] // Clear()
}

// NewRWMap returns a thread-safe map implementation
func NewRWMap[K comparable, V any]() (rwMap parli.ThreadSafeMap[K, V]) {
	return NewRWMap2[K, V]()
}

// NewRWMap2 returns a thread-safe map implementation
func NewRWMap2[K comparable, V any]() (rwMap *RWMap[K, V]) {
	return &RWMap[K, V]{threadSafeMap: *newThreadSafeMap[K, V]()}
}

// Putif is conditional Put depending on the return value from the putIf function.
//   - if key does not exist in the map, the put is carried out and wasNewKey is true
//   - if key exists and putIf is nil or returns true, the put is carried out and wasNewKey is false
//   - if key exists and putIf returns false, the put is not carried out and wasNewKey is false
//   - during PutIf, the map cannot be accessed and the map’s write-lock is held
//   - PutIf is an atomic, thread-safe operation
func (m *RWMap[K, V]) PutIf(key K, value V, putIf func(value V) (doPut bool)) (wasNewKey bool) {
	defer m.m2.Lock()()

	existing, keyExists := m.m2.Get(key)
	wasNewKey = !keyExists
	if keyExists && putIf != nil && !putIf(existing) {
		return // putIf false return: this value should not be updated
	}
	m.m2.Put(key, value)

	return
}

// Clone returns a shallow clone of the map
func (m *RWMap[K, V]) Clone() (clone parli.ThreadSafeMap[K, V]) {
	return m.Clone2()
}

// Clone returns a shallow clone of the map
func (m *RWMap[K, V]) Clone2() (clone *RWMap[K, V]) {
	defer m.m2.RLock()()

	return &RWMap[K, V]{threadSafeMap: *m.threadSafeMap.clone()}
}

// Keys provides the mapping keys, undefined ordering
//   - O(n)
//   - invoked while holding RLock or Lock
func (m *RWMap[K, V]) Keys(n ...int) (list []K) {
	// get n
	var n0 int
	if len(n) > 0 {
		n0 = n[0]
	}
	defer m.m2.RLock()()

	// handle n
	var length = m.m2.Length()
	if n0 == 0 {
		n0 = length
	} else if n0 > length {
		n0 = length
	}
	list = make([]K, n0)

	var r = ranger[K, V]{
		list: list,
		n:    n0,
	}
	m.m2.Range(r.rangeFunc)

	return
}

type ranger[K comparable, V any] struct {
	list []K
	i, n int
}

func (r *ranger[K, V]) rangeFunc(key K, value V) (keepGoing bool) {
	r.list[r.i] = key
	r.i++
	keepGoing = r.i < r.n
	return
}
