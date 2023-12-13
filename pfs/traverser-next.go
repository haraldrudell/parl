/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"io/fs"
	"path/filepath"
)

// Next returns the next file-system entry
//   - Next ends with [ResultEntry.DirEntry] nil, ie. [ResultEntry.IsEnd] returns true
//     or ResultEntry.Reason == REnd
//   - symlinks and directories can be skipped by invoking [ResultEntry.Skip].
//     Those have ResultEntry.Reason == RSkippable
//   - symlinks have information based on the symlink source but [ResultEntry.Abs] is the
//     fully resolved symlink target
//   - [ResultEntry.ProvidedPath] is a path to the entry based upon the initially
//     provided path. May be empty string
//   - [ResultEntry.Abs] is an absolute symlink-free clean path but only available when
//     [ResultEntry.Err] is nil
//   - [ResultEntry.Err] holds any error associated with the returned entry
//   - —
//   - result.Err is from:
//   - — process working directory cannot be read
//   - — directory read error or [os.Readlink] or [os.Lstat] failed
func (t *Traverser) Next() (result ResultEntry) {
	var entry ResultEntry

	// if pending initial path, create its root and entry
	if t.initialPath != "" {
		entry = t.createInitialRoot()
	} else {
		for {

			// process any pending returned entries
			if len(t.skippables) > 0 {
				entry = t.skippables[0]
				t.skippables[0] = ResultEntry{}
				t.skippables = t.skippables[1:]

				// handle skip and error
				if t.skipCheck(entry.No) {
					entry = ResultEntry{}
					continue // skip: not a directory that should be listed
				}

				// symlink that wasn’t skipped
				if entry.Type()&fs.ModeSymlink != 0 {
					t.processSymlink(entry.Abs)
					entry = ResultEntry{}
					continue // entry complete
				}

				// read directory that wasn’t skipped
				if entry.Err = t.readDir(entry.Abs, entry.ProvidedPath); entry.Err == nil {
					entry = ResultEntry{}
					continue // directory read successfully
				}
				entry.Reason = RDirBad
				// the directory is returned again for the error

				// process pending directory entries
			} else if len(t.dirEntries) > 0 {
				var dir = t.dirEntries[0]
				// number of directory entries typically ranges a dozen up to 3,000
				// one small slice alloc equals copy of 3,072 bytes: trimleft_bench_test.go
				// sizeof dirEntry is 48 bytes: 64 elements is 3,072 bytes
				// alloc is once per directory, copy is once per directory entry
				// number of copied elements for n is: [n(n+1)]/2: if 11 or more directory entries: alloc is faster
				// do alloc here
				t.dirEntries[0] = dirEntry{}
				t.dirEntries = t.dirEntries[1:]
				var name = dir.dirEntry.Name()
				entry.ProvidedPath = filepath.Join(dir.providedPath, name)
				entry.Abs = filepath.Join(dir.abs, name)
				entry.DirEntry = dir.dirEntry

				// process any additional roots
			} else {
				var root *Root2
				for t.rootIndex+1 < t.rootsRegistry.ListLength() {
					t.rootIndex++
					if root = t.rootsRegistry.GetValue(t.rootIndex); root != nil {
						break
					}
				}
				if root != nil {
					entry.ProvidedPath = root.ProvidedPath
					entry.Abs = root.Abs
					// either DirEntry or Err will be non-nil
					entry.DirEntry, entry.Err = AddDirEntry(entry.Abs)
				} else {

					// out of entries
					return // REnd
				}
			}

			// possibly return the entry
			//	- entry has ProvidedPath and (DirEntry/Abs or Err)
			//	- if entry is read from directory, IsDir/Type is always available
			//	- if entry.Err is non-nil, Abs and Name/IsDir/Type/Info are unavailable
			//	- entry may be any modeType including directory or symlink

			// resolve error-free symlinks
			if entry.Err == nil && entry.Type()&fs.ModeSymlink != 0 {
				// the symlink can be:
				//	- broken: entry.Err is non-nil
				//	- matching or a descendant of an existing root: ignore
				//	- a separate location creating a new root
				//	- a parent directory obsoleting an existing root

				// resolve all symlinks returning absolute, symlink-free, clean path
				var abs string
				if abs, entry.Err = filepath.EvalSymlinks(entry.Abs); entry.Err == nil {
					var dirEntry fs.DirEntry
					if dirEntry, entry.Err = AddDirEntry(abs); entry.Err == nil {
						// ProvidedPath is symlink source
						entry.Abs = abs
						// DirEntry is symlink source
						_ = dirEntry
						entry.Reason = RSkippable
						entry.No = t.skipNo.Add(1)
						t.skippables = append(t.skippables, entry)
					}
				}
				if entry.Err != nil {
					entry.Reason = RSymlinkBad
				}
			}

			// if entry is an obsolete root, it has already been traversed
			if entry.Err == nil && t.obsoleteRoots.HasAbs(entry.Abs) {
				entry = ResultEntry{}
				continue
			}

			break
		}
	}

	// the entry is to be returned
	if entry.Err == nil && entry.IsDir() {
		entry.No = t.skipNo.Add(1)
		t.skippables = append(t.skippables, entry)
	}
	result = entry
	if result.Reason == REnd {
		if result.Err == nil {
			if result.No != 0 {
				result.Reason = RSkippable
				result.SkipEntry = t.skip
			} else {
				result.Reason = REntry
			}
		} else {
			result.Reason = RError
		}
	}

	return // return of good or errored entry
}
