/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package swissmap

import (
	"iter"
	"unsafe"

	"github.com/haraldrudell/parl/plog"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	// H2 has is 7 bits, 128 combinations
	H2HashCombinations = 128
)

// [SwissMap] structure from runtime source code
//   - updated go1.24 250211
//   - a 64-bit hash is used
//   - GlobalDepth is the number of bits of the hash that is used to
//     identify table
//   - each table is a hash-map in itself with a number of mappings
//     and load factor
//
// [SwissMap]: https://github.com/golang/go/blob/master/src/internal/runtime/maps/map.go#L194C1-L242C1
type SwissMap struct {
	// The number of filled slots (i.e. the number of elements in all
	// tables). Excludes deleted slots.
	// Must be first (known by the compiler, for len() builtin).
	Used uint64

	// seed is the hash seed, computed as a unique random number per map.
	Seed uintptr

	// The directory of tables.
	//
	// Normally dirPtr points to an array of table pointers
	//
	// dirPtr *[dirLen]*table
	//
	// The length (dirLen) of this array is `1 << globalDepth`. Multiple
	// entries may point to the same table. See top-level comment for more
	// details.
	//
	// Small map optimization: if the map always contained
	// abi.SwissMapGroupSlots or fewer entries, it fits entirely in a
	// single group. In that case dirPtr points directly to a single group.
	//
	// dirPtr *group
	//
	// In this case, dirLen is 0. used counts the number of used slots in
	// the group. Note that small maps never have deleted slots (as there
	// is no probe sequence to maintain).
	DirPtr unsafe.Pointer
	DirLen int

	// The number of bits to use in table directory lookups.
	GlobalDepth uint8

	// The number of bits to shift out of the hash for directory lookups.
	// On 64-bit systems, this is 64 - globalDepth.
	GlobalShift uint8

	// writing is a flag that is toggled (XOR 1) while the map is being
	// written. Normally it is set to 1 when writing, but if there are
	// multiple concurrent writers, then toggling increases the probability
	// that both sides will detect the race.
	Writing uint8

	// clearSeq is a sequence counter of calls to Clear. It is used to
	// detect map clears during iteration.
	ClearSeq uint64
}

// GetSwissMap returns the Swiss map struct for Go map
//   - m: the map to examine. may be nil
//   - swissMap: a pointer to the live map structure
//   - Go versions prior to 1.24 returns nil
func GetSwissMap[K comparable, V any](m map[K]V) (swissMap *SwissMap) {

	// ensure go1.24+ and 64-bit
	if IsBucketMap() {
		return // too old Go version return swissMap == nil
	}

	// swissMap is pointer to the map’s live structure
	//	- the internal representation of a map value is pointer to a structure
	//	- swissMap may be nil
	swissMap = *(**SwissMap)(unsafe.Pointer(&m))

	return
}

// Groups iterates over the groups the map currently has allocated
//   - a group reference contains one or more groups
//   - a group contains SwissMapGroupSlots entries
func (m *SwissMap) Groups(yield func(group *SwissGroupsReference) (keepGoing bool)) {

	// if no directory: m.DirPtr is nil or points to single group
	if m.DirLen == 0 {
		if p := m.DirPtr; p != nil {
			var g = SwissGroupsReference{
				Data:       p,
				LengthMask: SwissMapGroupSlots - 1,
			}
			yield(&g)
		}
		return
	}

	// iterate over the directory
	//	- a table pointer may be shared across multiple slots
	//	- a table pointer may be nil
	var uniqifier = make(map[*SwissTable]struct{}, m.DirLen)
	for tableNo := range m.DirLen {
		var tablep = *(**SwissTable)(unsafe.Pointer(uintptr(m.DirPtr) + pruntime.PtrSize*uintptr(tableNo)))
		if tablep == nil {
			continue // nil pointer
		} else if _, exists := uniqifier[tablep]; exists {
			continue // re-used pointer
		}
		uniqifier[tablep] = struct{}{}
		if !yield(&tablep.Groups) {
			return
		}
	}
}

// SwissMap.Groups is iter.seq[*SwissGroupsReference]
var _ = func(m *SwissMap) (seq iter.Seq[*SwissGroupsReference]) { return m.Groups }

// Directory iterates over directory entries
//   - values are pointer to table
//   - table may be nil
//   - shares tables appears more than once
func (m *SwissMap) Directory(yield func(table *SwissTable) (keepGoing bool)) {
	for tableNo := range m.DirLen {
		var table = *(**SwissTable)(unsafe.Pointer(uintptr(m.DirPtr) + pruntime.PtrSize*uintptr(tableNo)))
		if !yield(table) {
			return
		}
	}
}

// SwissMap.Directory is iter.seq[*SwissTable]
var _ = func(m *SwissMap) (seq iter.Seq[*SwissTable]) { return m.Directory }

// Tables iterates over the tables of the directory
//   - table is never nil
//   - reused tables are uniqified and only returned once
func (m *SwissMap) Tables(yield func(table *SwissTable) (keepGoing bool)) {

	// iterate over the directory
	//	- a table pointer may be shared across multiple slots
	//	- a table pointer may be nil
	var uniqifier = make(map[*SwissTable]struct{}, m.DirLen)
	for table := range m.Directory {
		if table == nil {
			continue // nil pointer
		} else if _, exists := uniqifier[table]; exists {
			continue // re-used pointer
		}
		uniqifier[table] = struct{}{}
		if !yield(table) {
			return
		}
	}
}

// SwissMap.Tables is iter.seq[*SwissTable]
var _ = func(m *SwissMap) (seq iter.Seq[*SwissTable]) { return m.Tables }

// Structure returns the allocation state of the swiss map
//   - loadFactor is effective load factor across table and groups
//   - — zero if no allocations have been made or the map is empty
//   - — hash-map load factor is stored entries / number of buckets
//   - dirLen: size of directory 0…, power of 2
//   - — dirLen zero means there is one bucket if allocation has been made
//   - entryAllocationCount: the number of free or used mapping entries the map has currently allocated
//   - — 0…, across tables, multiple of 8
//   - groupCount: how many groups, the unit of allocation for swiss maps, that are allocated
//   - — 0…, each holding 8 entries
//     -
//   - swiss map structures grows independently and are lazily allocated
//   - therefore, each table has its own load factor
func (m *SwissMap) Structure() (loadFactor float32, dirLen, entryAllocationCount, tableCount, groupCount int) {

	// length of the directory 0…
	dirLen = m.DirLen

	// no directory: check for single group
	//	- m.DirPtr points to Group
	//	- the group can hold up to 8 entry-allocations
	if dirLen == 0 {

		// check for zero allocations
		if m.DirPtr == nil {
			return // zero allocations return: loadFactor dirLen entryAllocationCount, tableCount, groupCount zero
		}

		// one group case: bucket list length is one
		groupCount = 1
		entryAllocationCount = SwissMapGroupSlots
		// sequential comparison of each of 8 slots
		//	- integer comparison of lower 7-bits of the hash
		//	- key comparison for each matching 7-bit hash to verify collision or match
		//	- load factor is number of mappings divided by 128 hash slots
		//	- [getWithKeySmall]
		//
		// [getWithKeySmall]: https://github.com/golang/go/blob/master/src/internal/runtime/maps/map.go#L437
		var _ int
		if m.Used > 0 {
			loadFactor = float32(m.Used) / H2HashCombinations
		}
		return // one group return: dirLen tableCount zero, loadFactor possibly zero
	}

	// iterate over the directory
	//	- ignore nil tables
	//	- revisit reused tables therefore cannot use Tables method
	//	- each table has an equal likelyhood of being accessed
	var aggregateLoadFactor float32
	var uniqifier = make(map[*SwissTable]struct{}, m.DirLen)
	for table := range m.Directory {

		// ignore nil tables
		if table == nil {
			continue // nil pointer or empty table
		}

		// count unique tables and their allocations
		var isSharedTable bool
		if _, isSharedTable = uniqifier[table]; !isSharedTable {
			uniqifier[table] = struct{}{}
			if table.Capacity > 0 {
				entryAllocationCount += int(table.Capacity)
				groupCount += int(table.Capacity) / SwissMapGroupSlots
			}
		}
		if table.Capacity > 0 {
			var tableLoadFactor = float32(table.Used) / float32(table.Capacity)
			aggregateLoadFactor += tableLoadFactor
		}
	}
	loadFactor = aggregateLoadFactor / float32(m.DirLen)
	tableCount = len(uniqifier)

	return
}

// “swissMap_used_0_dirlen_0_depth_0
// _shift_0_dirptr_0x14000155c28_seq_0_seed_3226da43ac284c0a”
func (m *SwissMap) String() (s string) {
	return plog.EnglishSprintf(
		"swissMap_used_%d_dirlen_%d_depth_%d_shift_%d"+
			"_dirptr_0x%x_seq_%x_seed_%x",
		m.Used, m.DirLen, m.GlobalDepth, m.GlobalShift,
		m.DirPtr, m.ClearSeq, m.Seed,
	)
}
