/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// RWMap is a one-liner thread-safe mapping.
// RWMap implements [parli.ThreadSafeMap][K comparable, V any].
package pmaps

import (
	"github.com/haraldrudell/parl/parli"
	"github.com/haraldrudell/parl/perrors"
)

// RWMap is a one-liner thread-safe mapping.
// RWMap implements [parli.ThreadSafeMap][K comparable, V any].
//   - GetOrCreate method is an atomic, thread-safe operation as opposed to
//     Get-then-Put
//   - Swap and PutIf are atomic, thread-safe operations
//   - V is copied so if size of V is large or V contains locks, use pointer
//   - RWMap uses reader/writer mutual exclusion lock for slightly higher performance.
//   - RWMap must be in same package as PMap
//   - Get methods are O(1)
type RWMap[K comparable, V any] struct {
	ThreadSafeMap[K, V]
}

// NewRWMap returns a thread-safe map implementation
func NewRWMap[K comparable, V any]() (rwMap parli.ThreadSafeMap[K, V]) {
	return NewRWMap2[K, V]()
}

// NewRWMap2 returns a thread-safe map implementation
func NewRWMap2[K comparable, V any]() (rwMap *RWMap[K, V]) {
	return &RWMap[K, V]{ThreadSafeMap: *NewThreadSafeMap[K, V]()}
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
func (rw *RWMap[K, V]) GetOrCreate(
	key K,
	newV func() (value *V),
	makeV func() (value V),
) (value V, ok bool) {
	rw.lock.Lock()
	defer rw.lock.Unlock()

	// try existing mapping
	if value, ok = rw.m[key]; ok {
		return // mapping exists return
	}

	// create using newV
	if newV != nil {
		pt := newV()
		if pt == nil {
			panic(perrors.NewPF("newV returned nil"))
		}
		value = *pt
		rw.m[key] = value
		ok = true
		return // created using newV return
	}

	// create using makeV
	if makeV != nil {
		value = makeV()
		rw.m[key] = value
		ok = true
		return // created using makeV return
	}

	return // no key, no newV or makeV: nil return
}

// Putif is conditional Put depending on the return value from the putIf function.
//   - if key does not exist in the map, the put is carried out and wasNewKey is true
//   - if key exists and putIf is nil or returns true, the put is carried out and wasNewKey is false
//   - if key exists and putIf returns false, the put is not carried out and wasNewKey is false
//   - during PutIf, the map cannot be accessed and the map’s write-lock is held
//   - PutIf is an atomic, thread-safe operation
func (rw *RWMap[K, V]) PutIf(key K, value V, putIf func(value V) (doPut bool)) (wasNewKey bool) {
	rw.lock.Lock()
	defer rw.lock.Unlock()

	existing, keyExists := rw.m[key]
	wasNewKey = !keyExists
	if keyExists && putIf != nil && !putIf(existing) {
		return // putIf false return: this value should not be updated
	}
	rw.m[key] = value

	return
}

// Clone returns a shallow clone of the map
func (rw *RWMap[K, V]) Clone() (clone parli.ThreadSafeMap[K, V]) {
	return rw.Clone2()
}

// Clone returns a shallow clone of the map
func (rw *RWMap[K, V]) Clone2() (clone *RWMap[K, V]) {
	return &RWMap[K, V]{ThreadSafeMap: *rw.ThreadSafeMap.Clone()}
}

// Swap replaces the map with otherMap and returns the current map in previousMap
//   - if otherMap is not RWMap, no swap takes place and previousMap is nil
//   - Swap is an atomic, thread-safe operation
func (rw *RWMap[K, V]) Swap(otherMap parli.ThreadSafeMap[K, V]) (previousMap parli.ThreadSafeMap[K, V]) {

	// check otherMap
	replacingRWMap, ok := otherMap.(*RWMap[K, V])
	if !ok || replacingRWMap.m == nil {
		return // otherMap of bad type
	}

	// prepare previousMap
	replacedRWMap := &RWMap[K, V]{}
	previousMap = replacedRWMap

	rw.lock.Lock()
	defer rw.lock.Unlock()

	// swap
	replacedRWMap.m = rw.m
	rw.m = replacingRWMap.m

	return // swap complete return
}

// List provides the mapped values, undefined ordering
//   - O(n)
func (rw *RWMap[K, V]) List() (list []V) {
	rw.lock.RLock()
	defer rw.lock.RUnlock()

	list = make([]V, len(rw.m))
	i := 0
	for _, v := range rw.m {
		list[i] = v
		i++
	}

	return
}

// List provides keys, undefined ordering
//   - O(n)
func (rw *RWMap[K, V]) Keys(n ...int) (list []K) {

	// get n
	var n0 int
	if len(n) > 0 {
		n0 = n[0]
	}

	rw.lock.RLock()
	defer rw.lock.RUnlock()

	// get length
	var length int = len(rw.m)
	if n0 > 0 && n0 < length {
		length = n0
	}

	list = make([]K, length)
	i := 0
	for key := range rw.m {
		list[i] = key
		i++
		if i >= length {
			break
		}
	}

	return
}
