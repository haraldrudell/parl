/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

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

// ThreadSafeMap is a thread-safe reusable promotable Go map
//   - native Go map functions: Get Put Delete Length Range
//   - convenience functions: Clear Clone
//   - — those methods need access to the Go map
//   - lock control: Lock RLock
//   - ThreadSafeMap uses reader/writer mutual exclusion lock for thread-safety
type ThreadSafeMap[K comparable, V any] struct {
	lock sync.RWMutex
	m    map[K]V
}

// NewThreadSafeMap returns a thread-safe Go map
func NewThreadSafeMap[K comparable, V any]() (m *ThreadSafeMap[K, V]) {
	return &ThreadSafeMap[K, V]{m: make(map[K]V)}
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
//   - the ok return value is true if a mapping was found.
//   - invoked while holding Lock or RLock
//   - O(1)
func (m *ThreadSafeMap[K, V]) Get(key K) (value V, ok bool) {
	value, ok = m.m[key]
	return
}

// Put creates or replaces a mapping
//   - invoked while holding Lock
func (m *ThreadSafeMap[K, V]) Put(key K, value V) {
	m.m[key] = value
}

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
		delete(m.m, key)
		return // non-zero-value delete
	}

	// if key mapping does not exist: noop
	if _, itemExists := m.m[key]; !itemExists {
		return // write-free item does not exist return
	}

	// set value to zero to prevent temporary memory leaks
	var zeroValue V
	m.m[key] = zeroValue

	// delete
	delete(m.m, key)
}

// Length returns the number of mappings
//   - invoked while holding RLock or Lock
func (m *ThreadSafeMap[K, V]) Length() (length int) {
	return len(m.m)
}

// Range traverses map bindings
//   - iterates over map until rangeFunc returns false
//   - similar to: func (*sync.Map).Range(f func(key any, value any) bool)
//   - invoked while holding RLock or Lock
func (m *ThreadSafeMap[K, V]) Range(rangeFunc func(key K, value V) (keepGoing bool)) {
	for k, v := range m.m {
		if !rangeFunc(k, v) {
			return
		}
	}
}

// Clear empties the map
//   - if useRange is RangeDelete, the map is cleared by
//     iterating and deleteing all keys
//   - invoked while holding Lock
func (m *ThreadSafeMap[K, V]) Clear(useRange ...bool) {

	// if useRange is not present and true, clear by re-initialize
	if len(useRange) == 0 || !useRange[0] {
		m.m = make(map[K]V)
		return // re-create clear return
	}

	// zero-out and delete each item
	var zeroValue V
	for k := range m.m {
		m.m[k] = zeroValue
		delete(m.m, k)
	}
}

// Clone returns a shallow clone of the map
//   - clone is done by ranging all keys
//   - invoked while holding RLock or Lock
func (m *ThreadSafeMap[K, V]) Clone() (clone *ThreadSafeMap[K, V]) {
	return &ThreadSafeMap[K, V]{m: maps.Clone(m.m)}
}

// List provides the mapped values, undefined ordering
//   - O(n)
//   - invoked while holding RLock or Lock
func (m *ThreadSafeMap[K, V]) List(n int) (list []V) {

	// handle n
	var length = len(m.m)
	if n == 0 {
		n = length
	} else if n > length {
		n = length
	}

	// create and populate list
	list = make([]V, n)
	i := 0
	for _, v := range m.m {
		list[i] = v
		i++
		if i > n {
			break
		}
	}

	return
}

// invokes lock.Unlock()
func (m *ThreadSafeMap[K, V]) unlock() {
	m.lock.Unlock()
}

// invokes lock.RUnlock()
func (m *ThreadSafeMap[K, V]) runlock() {
	m.lock.RUnlock()
}
