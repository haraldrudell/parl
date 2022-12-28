/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "golang.org/x/exp/constraints"

// Ranking is a pointer-identity-to-value map of updatable values traversable by rank.
//   - V is a value reference composite type that is comparable, ie. not slice map function.
//     Preferrably, V is interface or pointer to struct type.
//   - R is an ordered type such as int, used to rank the V values
//   - values are added or updated using AddOrUpdate method distinguished by
//     (computer science) identity
//   - if the same comparable value V is added again, that value is re-ranked
//   - rank R is computed from a value V using the ranker function.
//     The ranker function may be examining field values of a struct
//   - values can have the same rank. If they do, equal rank is provided in insertion order
//   - pmaps.NewRanking[V comparable, R constraints.Ordered](
//     ranker func(value *V) (rank R))
//   - pmaps.NewRankingThreadSafe[V comparable, R constraints.Ordered](
//     ranker func(value *V) (rank R)))
type Ranking[V comparable, R constraints.Ordered] interface {
	// AddOrUpdate adds a new value to the ranking or updates the ranking of a value
	// that has changed.
	AddOrUpdate(value *V)
	// List returns the first n or default all values by rank
	List(n ...int) (rank []*V)
}

// ThreadSafeMap is a one-liner thread-safe mapping.
//   - pmaps.NewRWMap[K comparable, V any]()
type ThreadSafeMap[K comparable, V any] interface {
	// Get returns the value mapped by key or the V zero-value otherwise.
	//   - the ok return value is true if a mapping was found.
	//   - O(1)
	Get(key K) (value V, ok bool)
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
	GetOrCreate(
		key K,
		newV func() (value *V),
		makeV func() (value V),
	) (value V, ok bool)
	// Put saves or replaces a mapping
	Put(key K, value V)
	// Delete removes mapping using key K.
	//   - if key K is not mapped, the map is unchanged.
	//   - O(log n)
	Delete(key K)
	// Clear empties the map
	Clear()
	// Length returns the number of mappings
	Length() (length int)
	// Clone returns a shallow clone of the map
	Clone() (clone ThreadSafeMap[K, V])
	// List provides the mapped values, undefined ordering
	//   - O(n)
	List() (list []V)
}
