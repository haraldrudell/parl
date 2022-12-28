/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"strconv"
	"sync/atomic"
)

// UniqueID is an executable-invocation-unique identifier generator.
// The format is a named-type small-integer numeric string suitable to distinguish multiple instances of a type.
// The type is ordered and can be converted to string.
//
// Usage:
//
//	type MyType string
//	var generator parl.UniqueID[MyType]
//	someID := generator.ID()
type UniqueID[T ~string] uint64

// ID generates a unique string identifier. thread-safe
func (u *UniqueID[T]) ID() (unique T) {
	return T(strconv.FormatUint(atomic.AddUint64((*uint64)(u), 1), 10))
}
