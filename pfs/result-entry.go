/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"io/fs"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// ResultEntry is an existing file-system entry, traversal-end marker or an error-entry
// during file-system traversal
//   - typically created by [Traverser] or an iterator delegating to Traverser
//   - — “.” “..” are not returned
//   - non-error non-end entries have been proven to exist
//   - [ResultEntry.IsEnd] or Reason == REnd indicates end
//   - Reason indicates why the ResultEntry was returned:
//   - — [REnd] (zero-value) marks iteration ended: all fields are nil
//   - — — for-statement iterators do not return this value
//   - — [REntry] a directory entry: non-symlink non-directory non-error
//   - — [RSkippable] a symlink or directory that may be skipped, default is to traverse it
//   - — [RDirBad] a directory whose listing failed [os.Open] [os.File.ReadDir]
//   - — — such erroring directory is returned twice: RSkippable and RDirBad
//   - — — Err is set
//   - — [RSymlinkBad] a broken symlink, Err is set
//   - — [RError] error other than bad directory or bad symbolic link
//   - ResultEntry.Type().IsRegular() determines regular file
//   - — not available for [REnd]
//   - —
//   - ResultEntry is a value-container that as a local variable, function argument or result
//     is a tuple not causing allocation by using temporary stack storage
//   - taking the address of a &ResultEntry causes allocation
type ResultEntry struct {
	// Info() should be cached because it may be recreated on each invocation
	//	- — deferred-info executing [os.Lstat]
	//	- if Err is non-nil, [RError], Info may return error
	//	- nil for [REnd]
	fs.DirEntry
	// path as provided that may be easier to read
	//   - may be implicitly relative to current directory: “subdir/file.txt”
	//   - may have no directory or extension part: “README.css” “z”
	//   - may be relative: “../file.txt”
	//   - may contain symlinks, unnecessary “.” and “..” or
	//		multiple separators in sequence
	//	- may be empty string for current working directory
	//	- empty string for [REnd]
	ProvidedPath string
	//	- equivalent of Path: absolute, symlink-free, clean
	//	- if Err non-nil, [RError] may be empty
	//	- if the entry is symbolic link:
	//	- — ProvidedPath is the symbolic link location
	//	- — Abs is what the symbolic link points to
	//	- empty string for [REnd]
	Abs string
	// function to skip descending into directory or following symlink
	//	- only non-nil for directories and good symbolic links
	SkipEntry func(no uint64)
	// skippable serial number, argument to SkipEntry
	//	- only non-zero for directories and good symbolic links
	No uint64
	// any error associated with this entry
	//	- [os.Lstat] failed
	//	- only non-nil for [RDirBad] [RSymlinkBad] [RError]
	Err error
	// why this entry was provided
	//	- [REnd] [REntry] [RSkippable] [RDirBad] [RSymlinkBad] [RError]
	Reason ResultReason
}

// Skip marks the returned entry to be skipped
//   - the entry is directory or symbolic link
//   - can only be invoked when [ResultEntry.Reason] is [RSkippable]
func (e ResultEntry) Skip() { e.SkipEntry(e.No) }

// IsEnd indicates that this ResultEntry is an end of entries marker
func (e ResultEntry) IsEnd() (isEnd bool) { return e.DirEntry == nil }

// IsHidden indicates a hidden file, directory or other entry
func (e ResultEntry) IsHidden() (isHidden bool) {
	if e.DirEntry != nil {
		if n := e.Name(); n != "" {
			isHidden = n[0] == '.'
		}
	}
	return
}

// resultEntry path: "…" abs"/…" err ‘OK’ reason ‘non-symlink non-directory non-error entry’
func (e ResultEntry) String() (s string) {
	return parl.Sprintf("resultEntry path: %q abs%q err ‘%s’ reason ‘%s’",
		e.ProvidedPath, e.Abs,
		perrors.Short(e.Err),
		e.Reason,
	)
}
