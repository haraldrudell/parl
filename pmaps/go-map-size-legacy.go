/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"math"
	"unsafe"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pmaps/swissmap"
)

const (
	// a fake size for re-creating map structure
	// sizeof(int) is max 8, then there are two more uint8
	hmapFakeSizeInBytes = 10
	// offset to the B field of hmap
	hmapBOffset = int(unsafe.Sizeof(int(0))) + 1
)

// LegacyGoMapSize returns map bucket size for Go versions prior to go1.24
// or 32-bit
//   - for nil map or if hash-map is not used, size is zero
//   - hash map is used with Go prior to go1.24 and for 32-bit
//
// GoMapSize returns the current size of the bucket array of Go map m
//   - size is 0 for a nil map
//   - size is 1 for an unallocated hash-table — rare case
//   - otherwise size is a power of 2
//
// About Go map:
//   - Go map is a hash map
//   - a hash table is a space-time trade-off compared to array access
//   - size is how many slots m’s hash table currently has
//   - size may grow or shrink as m is modified
//   - a mapping of the hash value-space is used for hash-table array access
//   - each map slot contains a linked list of key-value pairs
//   - more slots is faster closing in on O(1) complexity, fewer slots saves memory
//   - Load factor is number of hash-table entries including collisions divided by
//     hash table size
//
// Source code:
//   - the map source code part of the runtime package is available online:
//   - — https://go.googlesource.com/go/+/refs/heads/master/src/runtime/map.go
//   - runtime source is typically installed on a computer that has Go:
//   - — module directory: …libexec/src, package directory: runtime
//   - — on macOS homebrew similar to: …/homebrew/Cellar/go/1.20.2/libexec/src
//
// prior to go1.24, map implementation was hmap map header,
// an array of buckets with a B uint8 2log number of buckets.
//   - was in file [runtime.hmap.go]
//
// delaration:
//
//	type hmap struct {
//	 	// Note: the format of the hmap is also encoded in cmd/compile/internal/gc/reflect.go.
//	 	// Make sure this stays in sync with the compiler's definition.
//	 	count     int // # live cells == size of map.  Must be first (used by len() builtin)
//	 	flags     uint8
//	 	B         uint8  // log_2 of # of buckets (can hold up to loadFactor * 2^B items)
//
// [runtime.hmap.go]: https://github.com/golang/go/blob/master/src/runtime/map.go
func LegacyGoMapSize[K comparable, V any](m map[K]V) (size int) {

	// filter for hashmap in use
	if m == nil || !swissmap.IsBucketMap() {
		return
	}

	// hmap is a pointer to runtime.hmap struct
	//   - hmap begins with an int which is count of elements in map 4/8/bytes
	//   - then there is uint8 which is flags
	//   - then there is uint8 B wich is 2log(hash-table size)
	var hmapp = *(**[hmapFakeSizeInBytes]uint8)(unsafe.Pointer(&m))

	// B is log2(hash-table size), uint8: 0…255
	//   - 2^255 ≈ 5e76, 1 GiB ≈ 1e9
	var B = (*hmapp)[hmapBOffset]
	if B > 63 { // B will not fit uint64
		panic(perrors.ErrorfPF("hash table size corrupt: 2^%d", B))
	}

	var size0 = uint64(1) << B // size = 2^B
	if size0 > uint64(math.MaxInt) {
		panic(perrors.ErrorfPF("hash table too large for int %d > %d", size0, math.MaxInt))
	}

	return
}
