/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"sync/atomic"

	"github.com/haraldrudell/parl"
)

// goerIndex is a one-based, thread-safe, panic-free generator
// of per-instance-unique parl.GoIndex values, ie. int.
// goIndex values can be used as map index.
// goerIndex does not require initialization.
type goerIndex struct {
	index uint64 // atomic
}

// goIndex returns the next parl.GoIndex value.
func (gi *goerIndex) goIndex() (goIndex parl.GoIndex) {
	return parl.GoIndex(atomic.AddUint64(&gi.index, 1))
}
