/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// goid.GoID() provides a unique goroutine identifier.
//
//	m := map[goid.ThreadID]SomeInterface{}
//	m[goid.GoID()] = …
package goid

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pdebug"
	"github.com/haraldrudell/parl/pruntime"
)

// GoID obtains a numeric string that as of Go1.18 is
// assigned to each goroutine. This number is an increasing
// unsigned integer beginning at 1 for the main invocation
func GoID() (threadID parl.ThreadID) {
	var err error
	if threadID, _, err = pdebug.ParseFirstLine(pruntime.FirstStackLine()); err != nil {
		panic(err)
	}
	return
}
