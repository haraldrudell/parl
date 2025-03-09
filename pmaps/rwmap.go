/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"github.com/haraldrudell/parl/parli"
	"github.com/haraldrudell/parl/pmaps/pmaps2"
)

// RWMap is a thread-safe mapping based on read/write mutex
//   - 5 native Go map functions: Get Put Delete Length Range
//   - — Delete optionally writes zero-value
//   - complex atomic methods:
//   - — GetOrCreate method is an atomic, thread-safe operation
//     as opposed to Get-then-Put
//   - — PutIf is atomic, thread-safe operation
//   - convenience methods:
//   - — Clone Clone2 based on [maps.Clone]
//   - — Clear using fast recreate or [maps.Range] optionally writing zero-values
//   - order functions:
//   - — List unordered values
//   - — Keys unordered keys
//   - V is copied so if V is large or contains locks, use pointer to V type
//   - RWMap implements [parli.ThreadSafeMap][K comparable, V any]
//   - —
//   - map mechanic is Go map
//   - RWMap uses reader/writer mutual exclusion lock for slightly higher performance
//   - Get methods are O(1)
//   - innermost type provides thread-safety
//   - outermost type provides map api
type RWMap[K comparable, V any] struct {
	// Get() GetOrCreate() Put() Delete() Length() Range() Clear() List()
	threadSafeMap[K, V]
}

// RWMap is parli.ThreadSafeMap
var _ parli.ThreadSafeMap[int, string] = &RWMap[int, string]{}

// NewRWMap returns a thread-safe map implementation as interface
func NewRWMap[K comparable, V any]() (rwMap parli.ThreadSafeMap[K, V]) {
	return NewRWMap2[K, V]()
}

// NewRWMap2 returns a thread-safe map implementation as pointer to type
func NewRWMap2[K comparable, V any](fieldp ...*RWMap[K, V]) (rwMap *RWMap[K, V]) {

	// set rwMap
	if len(fieldp) > 0 {
		rwMap = fieldp[0]
	}
	if rwMap == nil {
		rwMap = &RWMap[K, V]{}
	}

	// initialize all fields
	newThreadSafeMap(&rwMap.threadSafeMap)

	return
}

// Putif is conditional Put depending on the return value from the putIf function
//   - if key does not exist in the map, the put is carried out and wasNewKey is true
//   - if key exists and putIf is nil or returns true, the put is carried out and wasNewKey is false
//   - if key exists and putIf returns false, the put is not carried out and wasNewKey is false
//   - during PutIf, the map cannot be accessed and the map’s write-lock is held
//   - PutIf is an atomic, thread-safe operation
func (m *RWMap[K, V]) PutIf(key K, value V, putIf func(value V) (doPut bool)) (wasNewKey bool) {
	defer m.tsm.Lock().Unlock()

	existing, keyExists := m.tsm.Get(key)
	wasNewKey = !keyExists
	if keyExists && putIf != nil && !putIf(existing) {
		return // putIf false return: this value should not be updated
	}
	m.tsm.Put(key, value)

	return
}

// Clone returns a shallow clone of the map as interface type
func (m *RWMap[K, V]) Clone(goMap ...*map[K]V) (clone parli.ThreadSafeMap[K, V]) {
	return m.Clone2(goMap...)
}

// Clone returns a shallow clone of the map as implementation pointer type
func (m *RWMap[K, V]) Clone2(goMap ...*map[K]V) (clone *RWMap[K, V]) {
	var gm *map[K]V
	if len(goMap) > 0 {
		gm = goMap[0]
	}
	if gm == nil {
		clone = &RWMap[K, V]{}
	}
	defer m.tsm.RLock().RUnlock()

	// clone to Go-map case
	if gm != nil {
		m.threadSafeMap.cloneToGomap(gm)
		return
	}

	// clone RWMap case
	m.threadSafeMap.clone(&clone.threadSafeMap)

	return
}

// Keys provides the mapping keys, undefined ordering
//   - O(n)
//   - invoked while holding RLock or Lock
func (m *RWMap[K, V]) Keys(n ...int) (keys []K) {
	// get n
	var n0 int
	if len(n) > 0 {
		n0 = n[0]
	}
	defer m.tsm.RLock().RUnlock()

	// n0 is actual length to use
	var length = m.tsm.Length()
	if n0 <= 0 {
		n0 = length
	} else if n0 > length {
		n0 = length
	}

	var keyRanger = pmaps2.KeyRanger[K, V]{List: make([]K, n0)}
	m.tsm.Range(keyRanger.RangeFunc)
	keys = keyRanger.List

	return
}
