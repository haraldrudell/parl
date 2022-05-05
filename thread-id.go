/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

/*
ThreadID is an opaque type that uniquley identifies a thread,
ie. a goroutine.
goid.GoID obtains ThreadID for the executing
thread.
ThreadID is comparable, ie. can be used as a map key.
ThreadID can be cast to string using .String()
func (threadID ThreadID) String() (s string)
*/
type ThreadID string

func (threadID ThreadID) String() (s string) {
	return string(threadID)
}
