/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "strconv"

// ThreadID is an opaque type that uniquley identifies a thread, ie. a goroutine.
//   - goid.GoID obtains ThreadID for the executing thread.
//   - in runtime.g, goid is uint64
//   - ThreadID is comparable, ie. can be used as a map key.
//   - ThreadID is fmt.Stringer
//   - ThreadID has IsValid method
type ThreadID uint64

func (threadID ThreadID) IsValid() (isValid bool) {
	return threadID > 0
}

func (threadID ThreadID) String() (s string) {
	return strconv.FormatUint(uint64(threadID), 10)
}
