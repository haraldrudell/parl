/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"path/filepath"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// EntryScanner scans file-system entries and their children
type EntryScanner struct {
	symlinkCb                 func(abs string)
	firstPending, lastPending *PendingEntry
	pFSEntryCount             *int
}

func NewEntryScanner(
	rootEntry FSEntry, abs, path string,
	symlinkCb func(abs string),
	pFSEntryCount *int,
) (scanner *EntryScanner) {
	var pending = NewPendingEntry(abs, path, rootEntry)
	return &EntryScanner{
		symlinkCb:     symlinkCb,
		firstPending:  pending,
		lastPending:   pending,
		pFSEntryCount: pFSEntryCount,
	}
}

// Scan scans the file system for this root
//   - files and directories that are symlinks are stored via symlinkCb
//   - sub-directories have deferred processing using a linked list
func (s *EntryScanner) Scan() (err error) {

	// a linked list contains a pending root item and
	// pending sub-directories
	for s.firstPending != nil {

		// fetch abs path and entry for any pending item
		var abs = s.firstPending.Abs
		var path = s.firstPending.Path
		var entry = s.firstPending.Entry
		if s.firstPending = s.firstPending.Next; s.firstPending == nil {
			s.lastPending = nil
		}

		// process symlink for sub-directory or root entry
		if entry.IsSymlink() {
			s.symlinkCb(abs)
		}

		// file entry is now done
		if !entry.IsDir() {
			continue // file entry: go to next item
		}

		if parl.IsThisDebug() {
			St("scanner dir: %q abs: %q", path, abs)
		}

		// directory case: process children
		var dirEntry = entry.(*Directory)
		var children []FSEntry
		var paths []string
		// errors are stored in dirEntry or child
		children, paths, _ = dirEntry.FetchChildren(path)
		*s.pFSEntryCount += len(children)
		for i, child := range children {

			// obtain abs and path for child
			var p = paths[i]
			var a string
			if a, err = filepath.EvalSymlinks(filepath.Join(abs, child.Name())); perrors.Is(&err, "EvalSymlinks %w", err) {
				// likely encountered a broken symlink
				//	- scanning cannot stop
				//	- store error in child and continue
				if parl.IsThisDebug() {
					St("eval: abs: %q name: %q err: %q", abs, child.Name(), perrors.Short(err))
				}
				child.SetError(err)
				continue
			}

			// subdirectories are put in pending list
			if child.IsDir() {
				var pending = NewPendingEntry(a, p, child)
				if s.lastPending != nil {
					s.lastPending.Next = pending
				} else {
					s.firstPending = pending
				}
				s.lastPending = pending
			}

			// process symlink for file in a directory
			if entry.IsSymlink() {
				s.symlinkCb(abs)
			}
		}
	}

	return
}
