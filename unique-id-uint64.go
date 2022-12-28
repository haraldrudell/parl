/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"strconv"
	"sync/atomic"
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
type UniqueIDUint64 uint64

// ID generates a unique uint64 1…. thread-safe
func (u *UniqueIDUint64) ID() (uniqueUint64 uint64) {
	return atomic.AddUint64((*uint64)(u), 1)
}

func (u *UniqueIDUint64) String() (s string) {
	return "IDu64:" + strconv.FormatUint(uint64(*u), 10)
}
