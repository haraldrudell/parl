/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// RWMap is a one-liner thread-safe mapping.
// RWMap implements [parl.ThreadSafeMap][K comparable, V any].
package pmaps

import (
	"sync"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/maps"
)

// RWMap is a one-liner thread-safe mapping.
// RWMap implements [parl.ThreadSafeMap][K comparable, V any].
//   - GetOrCreate method is a thread-safe atomic operation as opposed to
//     Get-then-Put
//   - RWMap does not need to be initialized
//   - For using RWMap as periodically updated thread-safe mapping collection, use with
//     parl.AtomicReference
//   - V is copied so if size of V is large or V contains locks, use pointer
//   - RWMap uses reader/writer mutual exclusion lock for slightly higher performance.
//   - Get methods are O(1)
type RWMap[K comparable, V any] struct {
	lock sync.RWMutex
	m    map[K]V
}

// NewRWMap returns a thread-safe map implementation
func NewRWMap[K comparable, V any]() (rwMap parl.ThreadSafeMap[K, V]) {
	return &RWMap[K, V]{m: map[K]V{}}
}

// Get returns the value mapped by key or the V zero-value otherwise.
//   - the ok return value is true if a mapping was found.
//   - O(1)
func (rw *RWMap[K, V]) Get(key K) (value V, ok bool) {
	rw.lock.RLock()
	defer rw.lock.RUnlock()
	if rw.m == nil {
		rw.m = map[K]V{}
	}

	value, ok = rw.m[key]
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
//   - value insert is O(log n)
func (rw *RWMap[K, V]) GetOrCreate(
	key K,
	newV func() (value *V),
	makeV func() (value V),
) (value V, ok bool) {
	rw.lock.Lock()
	defer rw.lock.Unlock()
	if rw.m == nil {
		rw.m = map[K]V{}
	}

	// try exiting mapping
	if value, ok = rw.m[key]; ok {
		return // mapping exists return
	}

	// check if create
	if newV == nil && makeV == nil {
		return // no key, no newV or makeV: nil return
	}

	// create
	if newV != nil {
		pt := newV()
		if pt == nil {
			panic(perrors.New("Ordered newV: returned nil"))
		}
		rw.m[key] = *pt
	} else {
		rw.m[key] = makeV()
	}

	return
}

// Put saves or replaces a mapping
func (rw *RWMap[K, V]) Put(key K, value V) {
	rw.lock.Lock()
	defer rw.lock.Unlock()
	if rw.m == nil {
		rw.m = map[K]V{}
	}

	rw.m[key] = value
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (rw *RWMap[K, V]) Delete(key K) {
	rw.lock.Lock()
	defer rw.lock.Unlock()
	if rw.m == nil {
		rw.m = map[K]V{}
	}

	delete(rw.m, key)
}

// Clear empties the map
func (rw *RWMap[K, V]) Clear() {
	rw.lock.Lock()
	defer rw.lock.Unlock()
	if rw.m == nil {
		rw.m = map[K]V{}
	}

	maps.Clear(rw.m)
}

// Length returns the number of mappings
func (rw *RWMap[K, V]) Length() (length int) {
	rw.lock.Lock()
	defer rw.lock.Unlock()
	if rw.m == nil {
		rw.m = map[K]V{}
	}

	return len(rw.m)
}

// Clone returns a shallow clone of the map
func (rw *RWMap[K, V]) Clone() (clone parl.ThreadSafeMap[K, V]) {
	rw.lock.Lock()
	defer rw.lock.Unlock()

	var c RWMap[K, V]
	if rw.m != nil {
		c.m = maps.Clone(rw.m)
	}
	return &c
}

// List provides the mapped values, undefined ordering
//   - O(n)
func (rw *RWMap[K, V]) List() (list []V) {
	rw.lock.Lock()
	defer rw.lock.Unlock()
	if rw.m == nil {
		rw.m = map[K]V{}
	}

	list = make([]V, len(rw.m))
	i := 0
	for _, v := range rw.m {
		list[i] = v
		i++
	}

	return
}
