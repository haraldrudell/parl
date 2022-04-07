/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package goid

import "runtime/debug"

// GoRoutineID obtains a numeric string that as of Go1.18 is
// assigned to each goroutine. This number is an increasing
// unsigned integer beginning at 1 for the main invocation
func GoRoutineID() (threadID ThreadID) {
	var err error
	if threadID, _, err = ParseFirstStackLine(string(debug.Stack()), true); err != nil {
		panic(err)
	}
	return
}
