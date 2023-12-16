/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omaps

import (
	"github.com/google/btree"
	"github.com/haraldrudell/parl/perrors"
)

// ThreadSafeOrderedMapFunc is a mapping whose
// values are provided in custom order. Thread-safe.
//   - mapping implementation is Go Map
//   - native Go map functions: Get Put Delete Length Range
//   - convenience methods: Clone Clear
//   - order methods: List
//   - ordering structure is B-tree
//   - B-tree offers:
//   - — avoiding vector-copy of large sorted slices which is slow and
//   - — avoiding linear traversal of linked-lists which is slow and
//   - — is a more efficient structure than binary tree
type ThreadSafeOrderedMapFunc[K comparable, V any] struct {
	threadSafeMap[K, V] // Get() Length() Range()
	tree                *btree.BTreeG[V]
	less                btree.LessFunc[V]
}

// NewThreadSafeOrderedMapFunc returns a mapping whose
// values are provided in custom order. Thread-safe.
func NewThreadSafeOrderedMapFunc[K comparable, V any](
	less func(a, b V) (aBeforeB bool),
) (orderedMap *ThreadSafeOrderedMapFunc[K, V]) {
	return &ThreadSafeOrderedMapFunc[K, V]{
		threadSafeMap: *newThreadSafeMap[K, V](),
		tree:          btree.NewG(BtreeDegree, less),
		less:          less,
	}
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
) (value V, hasValue bool) {
	defer m.m2.Lock()()

	// try existing mapping
	//	- value is not comparable, so if mapping exists, the only
	//		action is to overwrite existing value
	if value, hasValue = m.m2.Get(key); hasValue {
		return // mapping exists return
	}

	// create using newV
	if hasValue = newV != nil; hasValue {
		if pt := newV(); pt == nil {
			panic(perrors.NewPF("newV returned nil"))
		} else {
			value = *pt
		}

		// create using makeV
	} else if hasValue = makeV != nil; hasValue {
		value = makeV()
	} else {

		return // no key, no newV or makeV: no value return
	}

	// store the created value
	m.m2.Put(key, value)
	m.tree.ReplaceOrInsert(value)

	return // value created return
}

func (m *ThreadSafeOrderedMapFunc[K, V]) Put(key K, value V) {
	defer m.m2.Lock()()

	// whether the mapping exists
	//	- value is not comparable, so if mapping exists, the only
	//		action is to overwrite existing value
	var _, hasExisting = m.m2.Get(key)

	// create or update mapping
	m.m2.Put(key, value)

	// if mapping exists, key is already in sort order
	if hasExisting {
		return // key already exists in order return
	}

	m.tree.ReplaceOrInsert(value)
}

func (m *ThreadSafeOrderedMapFunc[K, V]) Delete(key K, useZeroValue ...bool) {
	defer m.m2.Lock()()

	// no-op: delete non-existent mapping
	var existing, hasExisting = m.m2.Get(key)
	if !hasExisting {
		return // maping does not exist return
	}

	// delete mapping
	m.m2.Delete(key, useZeroValue...)
	m.tree.Delete(existing) // delete from sort order
}
func (m *ThreadSafeOrderedMapFunc[K, V]) Clone() (clone *ThreadSafeOrderedMapFunc[K, V]) {
	defer m.m2.RLock()()

	return &ThreadSafeOrderedMapFunc[K, V]{
		threadSafeMap: *m.threadSafeMap.clone(),
		tree:          m.tree.Clone(),
		less:          m.less,
	}
}
func (m *ThreadSafeOrderedMapFunc[K, V]) Clear(useRange ...bool) {
	defer m.m2.Lock()()

	m.m2.Clear(useRange...)
	m.tree.Clear(false)
}

// List provides mapped values in order
//   - n zero or missing means all items
//   - n non-zero means this many items capped by length
func (m *ThreadSafeOrderedMapFunc[K, V]) List(n ...int) (list []V) {
	defer m.m2.RLock()()

	// empty map case
	var length = m.m2.Length()
	if length == 0 {
		return
	}

	// non-zero list length [1..length] to use
	var nUse int
	// provided n capped by length
	if len(n) > 0 {
		if nUse = n[0]; nUse > length {
			nUse = length
		}
	}
	// default to full length
	if nUse == 0 {
		nUse = length
	}

	var err error
	if list, err = NewBtreeIterator[V, V](m.tree).Iterate(nUse); err != nil {
		panic(err)
	}

	return
}
