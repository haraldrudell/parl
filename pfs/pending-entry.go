/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

type PendingEntry struct {
	Next      *PendingEntry
	Abs, Path string
	Entry     FSEntry
}

func NewPendingEntry(abs, path string, entry FSEntry) (pending *PendingEntry) {
	return &PendingEntry{
		Abs:   abs,
		Path:  path,
		Entry: entry,
	}
}
