/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

const (
	Rfc3339s  = "2006-01-02 15:04:05Z07:00"
	Rfc3339ms = "2006-01-02 15:04:05.000Z07:00"
	Rfc3339us = "2006-01-02 15:04:05.000000Z07:00"
	Rfc3339ns = "2006-01-02 15:04:05.000000000Z07:00"
)

type FSLocation interface {
	Directory() (directory string)
}

// ThreadStatus indicates the current stat of a thread
// most often it is "running"
type ThreadStatus string
