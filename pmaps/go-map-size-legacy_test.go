/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"testing"

	"github.com/haraldrudell/parl/pmaps/pmaps2"
)

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
var _ int

func TestLegacyGoMapSize(t *testing.T) {

	if !pmaps2.IsBucketMap() {
		t.Skip("Go not featuring legacy bucket map")
	}

	const (
		// the hash table size of a nil map
		expNilSize = 0
		// size for making map
		sizeMoreThan2 = 3
	)
	var (
		size int
	)

	var m map[int]int

	// nil map
	if size = LegacyGoMapSize(m); size != expNilSize {
		t.Errorf("nil map size %d exp %d", size, expNilSize)
	}

	// map with more than 2 values
	m = make(map[int]int, sizeMoreThan2)
	if size = LegacyGoMapSize(m); size == expNilSize {
		t.Errorf("allocated map size %d exp %d", size, expNilSize)
	}
}
