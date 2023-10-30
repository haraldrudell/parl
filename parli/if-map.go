/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parli

const (
	MapDeleteWithZeroValue = true
	MapClearUsingRange     = true
)

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
	//	- newV and makeV may not access the map.
	//		The map’s write lock is held during their execution
	//	- GetOrCreate is an atomic, thread-safe operation
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
	// Delete removes mapping using key K.
	//   - if key K is not mapped, the map is unchanged.
	//   - O(log n)
	//	- if useZeroValue present and parli.MapDeleteWithZeroValue or true,
	//		the deleted value is fiorst assigned the zero-value
	Delete(key K, useZeroValue ...bool)
	// Clear empties the map
	//	- if useRange is present and parli.MapClearUsingRange or true,
	//		delete using range operations setting each item to zero-value
	//	- otherwise, clear by replacing with new map
	Clear(useRange ...bool)
	// Length returns the number of mappings
	Length() (length int)
	// Clone returns a shallow clone of the map
	Clone() (clone ThreadSafeMap[K, V])
	// List provides the mapped values, undefined ordering
	//   - O(n)
	List(n ...int) (list []V)
}
