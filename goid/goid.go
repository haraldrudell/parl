/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package goid

/*
ThreadID is an opaque type that uniquley identifies a thread,
ie. a goroutine.
pruntime.GoRoutineID obtains the ThreadID for the executing
thread.
ThreadID is comparable, ie. can be used as a map key.
ThreadID can be cast to string
*/
type ThreadID string

// ThreadStatus indicates the current stat of a thread
// most often it is "running"
type ThreadStatus string
