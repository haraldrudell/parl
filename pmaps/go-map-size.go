/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import "unsafe"

const (
	hmapFakeSizeInBytes = 10 // sizeof(int) is max 8, then there are two more uint8
	hmapBOffset         = int(unsafe.Sizeof(int(0))) + 1
)

// GoMapSize returns the current size of the bucket array of Go map m
//   - size is 0 for a nil map
//   - size is 1 if the hash-table is unallocated
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
//   - runtime source is typically installed on a computer that has go:
//   - — module directory: …libexec/src, package directory: runtime
//   - — on macOS homebrew similar to: …/homebrew/Cellar/go/1.20.2/libexec/src
func GoMapSize[K comparable, V any](m map[K]V) (size uint64) {

	// hmapp is a pointer to runtime.hmap struct
	//	- hmap begins with an int which is count of elements in map
	//	- then there is uint8 which is flags
	//	- then there is uint8 B wich is 2log(hash-table size)
	var hmapp *[hmapFakeSizeInBytes]uint8 = *(**[hmapFakeSizeInBytes]uint8)(unsafe.Pointer(&m))
	if hmapp == nil {
		return // nil map return
	}

	// B is log2(hash-table size), uint8: 0…255
	//	- 2^255 ≈ 5e76, 1 GiB ≈ 1e9
	var B = (*hmapp)[hmapBOffset]

	size = 1 << B // size = 2^B

	return
}
