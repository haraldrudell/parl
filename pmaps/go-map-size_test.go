/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"testing"
)

func TestGoMapSize(t *testing.T) {
	var nilSize = uint64(0)

	var size uint64

	var m map[int]int

	// nil map
	if size = GoMapSize(m); size != nilSize {
		t.Error("nil map not size 0")
	}

	// map with more than 2 values
	m = make(map[int]int, 3)
	if size = GoMapSize(m); size == nilSize {
		t.Error("map size 0")
	}
}
