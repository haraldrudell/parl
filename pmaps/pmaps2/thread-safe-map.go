/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps2

import (
	"sync"

	"golang.org/x/exp/maps"
)

const (
	// with [ThreadSafeMap.Delete] sets the mapping value to the
	// zero-value prior to delete
	SetZeroValue = true
	// with [ThreadSafeMap.Clear], the map is cleared using range
	// and delete of all keys rather than re-created
	RangeDelete = true
)

// ThreadSafeMap is a thread-safe reusable promotable hash-map
//   - ThreadSafeMap is the innermost type providing thread-safety to consuming
//     map implementations
//   - 5 native Go map functions: Get Put Delete Length Range
//   - — Delete optionally writes zero-value
//   - convenience methods:
//   - — Clone based on [maps.Clone]
//   - — Clear using fast recreate or [maps.Range] optionally writing zero-values
//   - lock control: Lock RLock
//   - order functions:
//   - — List unordered values
//   - ThreadSafeMap uses reader/writer mutual exclusion lock for thread-safety
//   - map mechnic is Go map
type ThreadSafeMap[K comparable, V any] struct {
	lock            *sync.RWMutex
	goMap           map[K]V
	unlock, runlock func()
}

// NewThreadSafeMap returns a thread-safe Go map
//   - stores self-refencing pointers
func NewThreadSafeMap[K comparable, V any]() (m *ThreadSafeMap[K, V]) {
	var rwm sync.RWMutex
	return &ThreadSafeMap[K, V]{
		lock:    &rwm,
		goMap:   make(map[K]V),
		unlock:  rwm.Unlock,
		runlock: rwm.RUnlock,
	}
}

// allows consumers to obtain the write lock
//   - returns a function releasing the lock
func (m *ThreadSafeMap[K, V]) Lock() (unlock func()) {
	m.lock.Lock()
	return m.unlock
}

// allows consumers to obtain the read lock
//   - returns a function releasing the lock
func (m *ThreadSafeMap[K, V]) RLock() (runlock func()) {
	m.lock.RLock()
	return m.runlock
}

// Get returns the value mapped by key or the V zero-value otherwise.
//   - hasValue is true if a mapping was found
//   - invoked while holding Lock or RLock
//   - O(1)
func (m *ThreadSafeMap[K, V]) Get(key K) (value V, hasValue bool) {
	value, hasValue = m.goMap[key]
	return
}

// Put creates or replaces a mapping
//   - invoked while holding Lock
func (m *ThreadSafeMap[K, V]) Put(key K, value V) { m.goMap[key] = value }

// Delete removes mapping for key
//   - if key is not mapped, the map is unchanged
//   - if useZeroValue is [pmaps.SetZeroValue], the mapping value is first
//     set to the zero-value. This prevents temporary memory leaks
//     when V contains pointers to large objects
//   - O(log n)
//   - invoked while holding Lock
func (m *ThreadSafeMap[K, V]) Delete(key K, useZeroValue ...bool) {

	// if doZero is not present and true, regular map delete
	if len(useZeroValue) == 0 || !useZeroValue[0] {
		delete(m.goMap, key)
		return // non-zero-value delete
	}

	// if key mapping does not exist: noop
	if _, itemExists := m.goMap[key]; !itemExists {
		return // write-free item does not exist return
	}

	// set value to zero to prevent temporary memory leaks
	var zeroValue V
	m.goMap[key] = zeroValue

	// delete
	delete(m.goMap, key)
}

// Length returns the number of mappings
//   - invoked while holding RLock or Lock
func (m *ThreadSafeMap[K, V]) Length() (length int) { return len(m.goMap) }

// Range traverses map bindings
//   - iterates over map until rangeFunc returns false
//   - similar to [sync.Map.Range] func (*sync.Map).Range(f func(key any, value any) bool)
//   - invoked while holding RLock or Lock
func (m *ThreadSafeMap[K, V]) Range(rangeFunc func(key K, value V) (keepGoing bool)) (rangedAll bool) {
	for k, v := range m.goMap {
		if !rangeFunc(k, v) {
			return
		}
	}
	return true
}

// Clear empties the map
//   - if useRange is RangeDelete, the map is cleared by
//     iterating and deleteing all keys
//   - invoked while holding Lock
func (m *ThreadSafeMap[K, V]) Clear(useRange ...bool) {

	// if useRange is not present and true, clear by re-initialize
	if len(useRange) == 0 || !useRange[0] {
		m.goMap = make(map[K]V)
		return // re-create clear return
	}

	// zero-out and delete each item
	var zeroValue V
	for k := range m.goMap {
		m.goMap[k] = zeroValue
		delete(m.goMap, k)
	}
}

// Clone returns a shallow clone of the map
//   - clone is done by ranging all keys
//   - invoked while holding RLock or Lock
func (m *ThreadSafeMap[K, V]) Clone() (clone *ThreadSafeMap[K, V]) {
	var rwm sync.RWMutex
	clone = &ThreadSafeMap[K, V]{
		lock:    &rwm,
		goMap:   maps.Clone(m.goMap),
		unlock:  rwm.Unlock,
		runlock: rwm.RUnlock,
	}

	return
}

// List provides the mapped values, undefined ordering
//   - O(n)
//   - invoked while holding RLock or Lock
func (m *ThreadSafeMap[K, V]) List(n int) (list []V) {

	// handle n
	var length = len(m.goMap)
	if n == 0 {
		n = length
	} else if n > length {
		n = length
	}

	// create and populate list
	list = make([]V, n)
	i := 0
	for _, v := range m.goMap {
		list[i] = v
		i++
		if i >= n {
			break
		}
	}

	return
}
