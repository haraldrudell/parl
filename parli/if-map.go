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
	//   - O(1) uses read lock
	Get(key K) (value V, ok bool)
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
	// Delete removes mapping using key K.
	//   - if key K is not mapped, the map is unchanged.
	//   - O(log n)
	//	- if useZeroValue present and parli.MapDeleteWithZeroValue or true,
	//		the deleted value is first assigned the zero-value.
	//		If V contains pointers, zero-value prevents temporary memory leaks
	Delete(key K, useZeroValue ...DeleteMethod)
	// Clear empties the map
	//	- if useRange is present and parli.MapClearUsingRange or true,
	//		delete using range operations setting each item to zero-value
	//	- otherwise, clear by replacing with new map
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

const (
	// with [ThreadSafeMap.Delete] sets the mapping value to the
	// zero-value prior to delete
	SetZeroValue DeleteMethod = true
	// with [ThreadSafeMap.Clear], the map is cleared using range
	// and delete of all keys rather than re-created
	RangeDelete ClearMethod = true
)

// whether map delete uses zero-value to avoid temporary memory leaks
//   - [SetZeroValue]
type DeleteMethod bool

// whether map clear uses range to clear the map
//   - [RangeDelete]
type ClearMethod bool
