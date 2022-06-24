/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"strconv"
	"sync/atomic"
)

// UniqueID is a executable-invocation-unique identifier generator.
// The identifier is a unique string. Usage:
//  type MyType string
//  var generator parl.UniqueID[MyType]
//  someID := generator.ID()
type UniqueID[T ~string] struct {
	lastID uint64
}

// ID generates a unique string identifier
func (u *UniqueID[T]) ID() (unique T) {
	return T(strconv.FormatUint(atomic.AddUint64(&u.lastID, 1), 10))
}
