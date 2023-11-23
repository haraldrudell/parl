/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// goid.GoID() provides a unique goroutine identifier.
package goid

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pdebug"
	"github.com/haraldrudell/parl/pruntime"
)

// GoID obtains a numeric string that as of Go1.18 is
// assigned to each goroutine
//   - [goid] 64-bit integer number incremented from
//     1 for the main invocation thread
//   - cache this value, it is expensive at 1.7 parallel mutex Lock/Unlock
//     via pruntime.FirstStackLine
//
// Usage:
//
//	m := map[goid.ThreadID]SomeInterface{}
//	cachedGoID := goid.GoID()
//	m[cachedGoID] = …
//
// [goid]: https://go.googlesource.com/go/+/go1.13/src/runtime/runtime2.go#409
func GoID() (threadID parl.ThreadID) {
	var err error
	if threadID, _, err = pdebug.ParseFirstLine(pruntime.FirstStackLine()); err != nil {
		panic(err)
	}
	return
}
