/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pmaps/swissmap"
)

// GoMapSize returns current properties for Go swiss map m
//   - m nil: loadFactor zero, allocatedEntries zero
//   - load factor: current value for the swiss map
//   - — because the swiss map hash bucket list is distributed over a
//     collection of tables, there is no unique hash bucket list size
//   - — load factor is calculated across all tables
//   - allocatedEntries: how many value-entry storage-locations currently allocated
//   - — swiss map features lazy allocation
//   - — allocation in multiples of 8
//   - — allocation size attempts to be a power of 2
//     -
//   - swiss map consists of a directory of tables that are themselves hash-maps,
//     and lists of groups that hold eight entry-allocations
//   - the tables grow independently and has their own load factor
//
// About go1.24 swiss map:
//   - Go map is an O(1) read hash-map
//   - Go map’s len function returns the number of currently assigned, undeleted mappings
//   - swiss map requries go1.24+ and 64-bit
//   - swiss map features multiple independently allocated table and group structures
//   - the O(1) hash function is provided by a directory of tables
//   - tables have their own load factor and grow independently
//
// About hash-table maps:
//   - a hash table is a space-time trade-off compared to array access
//   - a hash table is an associative array
//   - a mapping of the keys’ hashed value-space is used for hash-table array access
//   - the hash associates keys with value-slots in a list of buckets
//   - a hash collision is when multiple keys produce the same hash value
//   - each mapped slot contains one or more values, with a strategy to access multiple value-entries per slot
//   - the bucket-list length is practically always a power of 2
//   - a larger hash and bucket list is faster closing in on O(1) complexity, a shorter bucket list saves memory
//   - Load factor is the number of hash-table entries including collisions divided by the length of the bucket list
//
// Source code:
// [go1.24] introduces a map implementation based on [Swiss Tables]
//   - the map source code part of the runtime package is available online:
//   - — https://go.googlesource.com/go/+/refs/heads/master/src/runtime/map.go
//   - runtime source is typically installed on a computer that has Go:
//   - — module directory: …libexec/src, package directory: runtime
//   - — on macOS homebrew similar to: …/homebrew/Cellar/go/1.20.2/libexec/src
//
// prior to go1.24, map implementation was hmap map header, see [LegacyGoMapSize]
//
// [go1.24]: https://go.dev/blog/go1.24#performance-improvements
// [Swiss Tables]: https://abseil.io/about/design/swisstables
func GoMapSize[K comparable, V any](m map[K]V) (loadFactor float32, allocatedEntries int) {

	// ensure a swiss map structure to be available
	if m == nil || swissmap.IsBucketMap() {
		return // nil map or bucket-map return: loadFactor and allocatedEntries zero
	}

	// get swiss-map internal structure
	var swissMap = swissmap.GetSwissMap(m)
	// already checked, should never happen
	if swissMap == nil {
		panic(perrors.New("swissMap cannot be nil"))
	}

	// the map stores in groups of particular dimension
	//	- the map can have 0 or 1 groups
	//		or a directory of any number of groups
	//	- directory length is 2^globalDepth

	loadFactor, _ /*dirLen*/, allocatedEntries, _ /*tableCount*/, _ /*groupCount*/ = swissMap.Structure()

	return
}
