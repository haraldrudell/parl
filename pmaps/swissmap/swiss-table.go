/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package swissmap

// table is a Swiss table hash table structure [SwissTable]
//   - Each table is a complete hash table implementation.
//   - Map uses one or more tables to store entries. Extendible hashing (hash
//     prefix) is used to select the table to use for a specific key. Using
//     multiple tables enables incremental growth by growing only one table at a
//     time.
//   - updated go1.24 250211
//
// [SwissTable]: https://github.com/golang/go/blob/master/src/internal/runtime/maps/table.go#L34C6-L34C11
type SwissTable struct {
	// The number of filled slots (i.e. the number of elements in the table).
	Used uint16

	// The total number of slots (always 2^N). Equal to
	// `(groups.lengthMask+1)*abi.SwissMapGroupSlots`.
	Capacity uint16

	// The number of slots we can still fill without needing to rehash.
	//
	// We rehash when used + tombstones > loadFactor*capacity, including
	// tombstones so the table doesn't overfill with tombstones. This field
	// counts down remaining empty slots before the next rehash.
	GrowthLeft uint16

	// The number of bits used by directory lookups above this table. Note
	// that this may be less then globalDepth, if the directory has grown
	// but this table has not yet been split.
	LocalDepth uint8

	// Index of this table in the Map directory. This is the index of the
	// _first_ location in the directory. The table may occur in multiple
	// sequential indicies.
	//
	// index is -1 if the table is stale (no longer installed in the
	// directory).
	Index int

	// groups is an array of slot groups. Each group holds abi.SwissMapGroupSlots
	// key/elem slots and their control bytes. A table has a fixed size
	// groups array. The table is replaced (in rehash) when more space is
	// required.
	//
	// TODO(prattmic): keys and elements are interleaved to maximize
	// locality, but it comes at the expense of wasted space for some types
	// (consider uint8 key, uint64 element). Consider placing all keys
	// together in these cases to save space.
	Groups SwissGroupsReference
}
