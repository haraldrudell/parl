/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"strconv"
	"sync/atomic"
	"unsafe"

	"github.com/haraldrudell/parl/perrors"
)

// UniqueIDUint64 generates executable-invocation-unique uint64 numbers 1…
// Different series of uint64 from different generators does not have identity,
// ie. they cannot be told apart.
// Consider UniqueIDTypedUint64 to have identity.
//
// Usage:
//
//	var generator parl.UniqueIDUint64
//	id := generator.ID()
type UniqueIDint int

var intSize = int(unsafe.Sizeof(2))

// ID generates a unique uint64 1…. thread-safe
func (u *UniqueIDint) ID() (uniqueID int) {
	switch intSize {
	case 4:
		return int(atomic.AddInt32((*int32)(unsafe.Pointer(u)), 1))
	case 8:
		return int(atomic.AddInt64((*int64)(unsafe.Pointer(u)), 1))
	}
	panic(perrors.ErrorfPF("int size not 4 or 8: %d", intSize))
}

func (u *UniqueIDint) String() (s string) {
	return "IDint:" + strconv.FormatInt(int64(*u), 10)
}
