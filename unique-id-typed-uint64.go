/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"strconv"
	"sync/atomic"
)

// UniqueIDTypedUint64 generates integer-based opaque uniquely-typed IDs. Thread-safe.
//   - Different named-type series have their own unique type and can be told apart.
//   - The named types are ordered integer-based with String method implementing fmt.Stringer.
//
// Usage:
//
//	type T uint64
//	var generator parl.UniqueIDTypedUint64[T]
//	func (t T) String() string { return generator.StringT(t) }
//	someID := generator.ID()
type UniqueIDTypedUint64[T ~uint64] uint64

// ID generates a unique ID of integral type. Thread-safe
func (u *UniqueIDTypedUint64[T]) ID() (uniqueT T) {
	return T(atomic.AddUint64((*uint64)(u), 1))
}

// "parl.T:2", zero-value t prints "parl.T:0"
func (u *UniqueIDTypedUint64[T]) StringT(t T) (s string) {
	return Sprintf("%T:%s", t, strconv.FormatUint(uint64(t), 10))
}
