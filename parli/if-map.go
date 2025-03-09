/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parli

const (
	// [ThreadSafeMap.Delete] deleted mappings are first written with
	//	zero-value to prevent temporary memory leaks from pointers in value
	MapDeleteWithZeroValue DeleteMethod = true
)
const (
	// [ThreadSafeMap.Clear] clear the map by ranging all mappings
	//	- default is to clear the map by replacing it with a created map
	//	- — this prevents temporaty mwemory leak from large allocated
	//		structures
	MapClearUsingRange ClearMethod = true
)

// ThreadSafeMap is a one-liner thread-safe mapping
//   - implementation based on [sync.RWMutex] is
//     [github.com/haraldrudell/parl/pmaps.RWMap]
type ThreadSafeMap[K comparable, V any] interface {
	// Get returns the value mapped by key or the V zero-value otherwise
	//	- key: the mappin gkey whose value to retrieve
	//	- value: a V or V zero-value
	//	- hasValue true: mapping existed, value is valid
	//	- hasValue false: a mapping did not exist, value is zero-value
	//   - Go map O(1), RWMap uses read lock
	Get(key K) (value V, hasValue bool)
	// GetOrCreate returns an item from the map if it exists otherwise creates it.
	//   - if a key is mapped, its value is returned
	//   - otherwise, if newV and makeV are both nil, nil is returned.
	//   - otherwise, if newV is present, it is invoked to return a pointer to a value.
	//     A nil return value from newV causes panic. A new mapping is created using
	//     the value pointed to by the newV return value.
	//   - otherwise, a mapping is created using makeV return value
	//	- newV and makeV may not access the map.
	//		The map’s write lock is held during their execution
	//	- newV pointer facilitates saving a copy of existing value in the map using single copy operation
	//	- makeV facilitates saving value without allocation
	//	- GetOrCreate is an atomic, thread-safe operation
	//	- GetOrCreate uses write lock, not read lock like Get
	//   - value insert is O(log n)
	GetOrCreate(
		key K,
		newV func() (value *V),
		makeV func() (value V),
	) (value V, ok bool)
	// Put saves or replaces a mapping
	Put(key K, value V)
	// Putif is conditional Put depending on the return value from the putIf function.
	//	- if key does not exist in the map, the put is carried out and wasNewKey is true
	//	- if key exists and putIf is nil or returns true, the put is carried out and wasNewKey is false
	//	- if key exists and putIf returns false, the put is not carried out and wasNewKey is false
	//   - during PutIf, the map cannot be accessed and the map’s write-lock is held
	//	- PutIf is an atomic, thread-safe operation
	PutIf(key K, value V, putIf func(value V) (doPut bool)) (wasNewKey bool)
	// Delete removes mapping of key K
	//   - key: mapping key to  delete.
	//		If K is not mapped, the map is unchanged
	//	- useZeroValue [MapDeleteWithZeroValue]: the deleted value is first assigned the zero-value.
	//		If V contains pointers, zero-value prevents temporary memory leaks
	//	- —
	//   - Go map implementations O(log n)
	Delete(key K, useZeroValue ...DeleteMethod)
	// Clear empties the map
	//	- useRange [MapClearUsingRange]: delete using range operations setting each item to zero-value
	//	- useRange missing: clear by replacing with new map
	Clear(useRange ...ClearMethod)
	// Length returns the number of mappings
	Length() (length int)
	// Clone returns a shallow clone of the map
	//	- if goMap is present and non-nil, it receives a Go-map clone. return value is then nil
	Clone(goMap ...*map[K]V) (clone ThreadSafeMap[K, V])
	// List provides the mapped values, undefined ordering
	//   - O(n)
	List(n ...int) (list []V)
}

// ValueOrderedMap is a mapping whose
// values are provided in custom order
//   - implemented by [github.com/haraldrudell/parl/omaps/ValueOrderedMap]
type ValueOrderedMap[K comparable, V any] interface {
	// Get Put Delete Length Range
	GoMap[K, V]
	// Clear empties the map
	//   - may clear by re-creating the map
	//   - may clear by ranging and deleting all keys,
	//     retaining the allocated map-size
	Clear()
	// Clone returns a shallow clone of the map
	//	- goMap: optional pointer to receiving Go map instance
	//	- — Clone returns nil
	Clone(goMap ...*map[K]V) (clone ValueOrderedMap[K, V])
	// List provides mapped values in order
	//   - n zero, missing: list contains all items
	//   - n non-zero: list contains zero to n items capped by map length
	List(n ...int) (list []V)
}

// Map is a reusable promotable Go map
//   - 5 native Go Map functions: Get Put Delete Length Range
//   - 2 convenience functions: Clear Clone
//   - implemented by [github.com/haraldrudell/parl/pmaps/pmaps2/Map]
type Map[K comparable, V any] interface {
	// Get Put Delete Length Range
	GoMap[K, V]
	// Clear empties the map
	//   - may clear by re-creating the map
	//   - may clear by ranging and deleting all keys,
	//     retaining the allocated map-size
	Clear()
	// Clone returns a shallow clone of the map
	//	- goMap: optional pointer to receiving Go map instance
	//	- — Clone returns nil
	Clone(goMap ...*map[K]V) (clone Map[K, V])
}

// GoMap is a reusable promotable Go map
//   - 5 native Go Map functions: Get Put Delete Length Range
//   - GoMap is a base interface implemented as [Map] and other derived interfaces
type GoMap[K comparable, V any] interface {
	// Get returns a possibly mapped value mapped by key or the V zero-value otherwise
	//   - key: the key of the mapping
	//   - value: the value mapped or V zero-value
	//   - hasValue true: value is valid, the mapping did exist
	//   - hasValue false: value is zero-value, the mapping did not exist
	//   - amortized O(1) for Go map implementations
	Get(key K) (value V, hasValue bool)
	// Put creates or replaces a mapping
	Put(key K, value V)
	// Delete removes mapping for key
	//   - if key is not mapped, the map is unchanged
	//	- may zero-set the mapping to avoid temporary
	//		memory leaks from retained V values
	//   - O(log n) for Go map implementations
	Delete(key K)
	// Length returns the number of mappings
	Length() (length int)
	// Range traverses map bindings
	//   - iterates over map until rangeFunc returns false
	//   - order is undefined
	//   - similar to [sync.Map.Range] func (*sync.Map).Range(f func(key any, value any) bool)
	Range(rangeFunc func(key K, value V) (keepGoing bool)) (rangedAll bool)
}

// whether map delete uses zero-value to avoid temporary memory leaks
//   - [MapDeleteWithZeroValue]
type DeleteMethod bool

// whether map clear uses range to clear the map
//   - [MapClearUsingRange]
type ClearMethod bool
