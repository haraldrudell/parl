/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import "io/fs"

// ResultEntry is an entry, end or error during file-system traversal
//   - non-error non-end entries do exist
type ResultEntry struct {
	//	- always non-nil
	//	- may be deferred-info version, ie. only invoke Info once
	//	- if Err non-nil, Info may return error
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

func (e ResultEntry) Skip()               { e.SkipEntry(e.No) }
func (e ResultEntry) IsEnd() (isEnd bool) { return e.DirEntry == nil }
