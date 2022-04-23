/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package goid provides goid.GoID(), unique goroutine identifiers
//  m := map[goid.ThreadID]SomeInterface{}
//  m[goid.GoID()] = …
package goid

import (
	"runtime/debug"

	"github.com/haraldrudell/parl"
)

// GoID obtains a numeric string that as of Go1.18 is
// assigned to each goroutine. This number is an increasing
// unsigned integer beginning at 1 for the main invocation
func GoID() (threadID parl.ThreadID) {
	var err error
	if threadID, _, err = ParseFirstLine(string(debug.Stack())); err != nil {
		panic(err)
	}
	return
}
