/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import "fmt"

const (
	// the filesystem traversal completed all entries
	REnd ResultReason = iota
	// a non-symlink non-directory non-error entry
	REntry
	// a directory or symlink about to be traversed
	RSkippable
	// a directory whose listing failed [os.Open] [os.File.ReadDir]
	RDirBad
	// a broken symlink [os.Stat]
	RSymlinkBad
	// failure in [os.Lstat] [os.Readlink] [os.Getwd]
	RError
)

// Why a directory entry was provided by Traverser
//   - REnd REntry RSkippable RDirBad RSymlinkBad RError
type ResultReason uint8

var reasonMap = map[ResultReason]string{
	REnd:        "filesystem traversal completed all entries",
	REntry:      "non-symlink non-directory non-error entry",
	RSkippable:  "a directory or symlink about to be traversed",
	RDirBad:     "a directory whose listing failed",
	RSymlinkBad: "a broken symlink",
	RError:      "error while examining entry",
}

func (r ResultReason) String() (s string) {
	if s = reasonMap[r]; s != "" {
		return
	}
	s = fmt.Sprintf("?entryReason%d", r)

	return
}
