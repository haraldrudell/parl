/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import "io/fs"

// ResultEntry is an existing file-system entry, traversal-end marker or an error-entry
// during file-system traversal
//   - non-error non-end file-system entries have been proven to exist
//   - [ResultEntry.IsEnd] or Reason == REnd indicates end
//   - Reson indicates why the ResultEntry was returned:
//   - [REnd] [REntry] [RSkippable] [RDirBad] [RSymlinkBad] [RError]
//   - —
//   - ResultEntry is a value-container that as a local variable, function argument or result
//     is a tuple not causing allocation by using temporary stack storage
//   - taking the address of a &ResultEntry causes allocation
type ResultEntry struct {
	//	- always non-nil
	//	- may be deferred-info executing [os.Lstat] every time, ie. only invoke Info once
	//	- if Err is non-nil, Info may return error
	fs.DirEntry
	// path as provided that may be easier to read
	//   - may be implicitly relative to current directory: “subdir/file.txt”
	//   - may have no directory or extension part: “README.css” “z”
	//   - may be relative: “../file.txt”
	//   - may contain symlinks, unnecessary “.” and “..” or
	//		multiple separators in sequence
	//	- may be empty string for current working directory
	ProvidedPath string
	//	- equivalent of Path: absolute, symlink-free, clean
	//	- if Err non-nil, may be empty
	//	- if the entry is symbolic link:
	//	- — ProvidedPath is the symbolic link location
	//	- — Abs is what the symbolic link points to
	Abs string
	// function to skip descending into directory or following symlink
	SkipEntry func(no uint64)
	// skippable serial number
	No uint64
	// any error associated with this entry
	Err error
	// why this entry was provided
	Reason ResultReason
}

// Skip marks the returned entry to be skipped
//   - the entry is directory or symbolic link
//   - can only be invoked when [ResultEntry.Reason] is [RSkippable]
func (e ResultEntry) Skip() { e.SkipEntry(e.No) }

// IsEnd indicates that this ResultEntry is an end of entries marker
func (e ResultEntry) IsEnd() (isEnd bool) { return e.DirEntry == nil }
