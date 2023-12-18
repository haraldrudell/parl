/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package pfs provides a symlink-following file-systemtraverser and other file-system functions.
package pfs

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices"
)

const (
	// platform path separator as a string
	sSep = string(filepath.Separator)
	// [os.File.ReadDir] get all names
	allNamesAtOnce = -1
)

// Traverser represents a file system that is scanned following symlinks
//   - each file system entry is returned exactly once except:
//   - — error reading a directory returns the directory a second time
//   - directories and symlinks are returned before they are read so that they can be
//     more efficiently skipped by invoking [ResultEntry.Skip]
//   - directory entries are returned in 8-bit character order
//   - returned entries may not exist, such entries have [ResultEntry.Err] non-nil
//   - result.ProvidedPath is based on the initial path and may be relative,
//     unclean and contain symlinks
//   - if [ResultEntry.Err] is nil, Abs is absolute, symlink-free clean path
//   - —
//   - ResultEntry.DirEntry.Info typically invokes [os.Lstat] every time,
//     so this value should be cached
//   - because symlinks may point to parents or separate trees,
//     the file system scan may involve multiple roots which may
//     affect the order of return entries
//   - symlinks are followed and not returned.
//     Therefore, a symlink pointing to a scanned location is effectively ignored
//   - the returned struct is by value. If its address is not taken,
//     no allocation will occur
type Traverser struct {
	// path provided to new-function for the initial root
	initialPath string
	// skipNo provides a serial number for returned directories
	skipNo atomic.Uint64
	// skippables holds pending skippables
	skippables []ResultEntry
	// collection of skippables marked to be skipped
	skipMap map[uint64]struct{}
	// basenames from read directories to be processed
	dirEntries []dirEntry
	// index in rootsRegistry being traversed
	rootIndex int
	// registry of the absolute paths for each encountered root
	//	- key: absolute, symlink-free, clean path
	rootsRegistry Registry[Root2]
	// obsoleteRoots were obsoleted by a symlink pointing to
	//		a parent directory
	//	- these roots will be encountered during traversal
	obsoleteRoots Registry[Root2]
}

// dirEntry is a value-container for a read directory entry
//   - [os.File.ReadDir] returns dirEntry with deferred [fs.FileInfo]
type dirEntry struct {
	abs, providedPath string
	dirEntry          fs.DirEntry
}

// NewTraverser returns a file-system traverser
//   - typically used via [pfs.Iterator] or [pfs.DirIterator]
//   - path is the initial path.
//     Path may be relative or absolute, contain symlinks and be unclean.
//     Path may be of any modeType: file, directory or special file.
//     Empty string means process’ current directory
//   - the Next method is used to obtain file-system entries and errors
//   - consider using pfs iterators:
//   - — [Iterator] for all entries and errors
//   - — [DirIterator] for error-free directories
//
// Usage:
//
//	var traverser = pfs.NewTraverser(path)
//	for {
//	  var result = traverser.Next()
//	  if result.IsEnd() || result.Err != nil {
//	    break
//	  }
//	  println(result.Abs)
//	}
func NewTraverser(path string) (traverser *Traverser) {
	return &Traverser{
		initialPath:   path,
		skipMap:       make(map[uint64]struct{}),
		rootsRegistry: *NewRegistry[Root2](),
		obsoleteRoots: *NewRegistry[Root2](),
	}
}

// skip marks no for skipping
func (t *Traverser) skip(no uint64) { t.skipMap[no] = struct{}{} }

// skipCheck returns true if no is marked for skipping
func (t *Traverser) skipCheck(no uint64) (skip bool) {
	if _, skip = t.skipMap[no]; !skip {
		return
	}
	delete(t.skipMap, no)

	return
}

// createInitialRoot returns the createInitialRoot entry and creates and registers its root
//   - entry is non-nil and may be symbolic link
//   - entry has ProvidedPath and DirEntry
//   - if entry.Err is nil, Abs and Name/IsDir/Type/Info are available
func (t *Traverser) createInitialRoot() (entry ResultEntry) {

	// create a root for path provided to NewTree2
	var root = NewRoot2(t.initialPath)
	t.initialPath = ""

	// load absolute, symlink-free, clean path
	//	- errors if [os.Getwd] or [os.Readlink] fails
	if entry.Err = root.Load(); entry.Err == nil {
		// the root is usable
		t.rootsRegistry.Add(root.Abs, root)
	}

	var err error
	// modeType is required to examine the entry
	//	- it is not available, so [os.Lstat] and [os.Stat] must be invoked
	//	- start with Lstat to see if it is a symlink
	entry.ProvidedPath = root.ProvidedPath
	entry.Abs = root.Abs
	if entry.Abs != "" {
		entry.DirEntry, entry.Err = AddDirEntry(entry.Abs)
	} else {
		// provide best-effort DirEntry
		if entry.DirEntry, err = AddDirEntry(entry.ProvidedPath); err != nil {
			entry.Err = perrors.AppendError(entry.Err, err)
		}
	}

	// if Lstat failed, use a deferred-error dirEntry
	if entry.DirEntry == nil {
		entry.DirEntry = NewDeferringDirEntry(entry.ProvidedPath)
	}

	return
}

// processSymlink checks for new or obsoleted roots from a symlink
func (t *Traverser) processSymlink(absTarget string) {

	// check for exact match to existing root
	if t.rootsRegistry.HasAbs(absTarget) {
		return // symlink matches existing root: ignore it
	}

	// match absTarget against existing roots
	var length = t.rootsRegistry.ListLength()
	for i := 0; i < length; i++ {

		// iterate over roots
		var root = t.rootsRegistry.GetValue(i)
		if root == nil {
			continue // a discarded root
		}
		var rootAbs = root.Abs + sSep
		var targetAbs = absTarget + sSep

		// if absTarget is a subdirectory of an existing root, it can be ignored
		if strings.HasPrefix(targetAbs, rootAbs) {
			return // symlink is a sub-entry of an existing root: ignore it
		}

		// if root is not a subdirectory of absTarget, check the next root
		if !strings.HasPrefix(rootAbs, targetAbs) {
			continue
		}

		// root is a subdirectory of this symlink, obsolete the root
		if i <= t.rootIndex {
			// the obsolete root was already being traversed
			//	- save it
			t.obsoleteRoots.Add(root.Abs, root)
		}
		t.rootsRegistry.ObsoleteIndex(i)
	}
	// the symlink is disparate from all existing roots

	// scan as new root
	var root = NewAbsRoot2(absTarget)
	t.rootsRegistry.Add(absTarget, root)
}

// readDir reads a directory and adds entries to t.dirEntries
func (t *Traverser) readDir(abs, providedPath string) (err error) {

	// DirEntry with basename and modeType
	var entries []fs.DirEntry
	if entries, err = t.directoryOrder(abs); err != nil {
		return
	}

	// sort by 8-bit characters
	slices.SortFunc(entries, compareDirEntry)

	// create entries for Next function
	//	- defers symlink resolution
	var index, endIndex = len(t.dirEntries), len(t.dirEntries) + len(entries)
	pslices.SetLength(&t.dirEntries, endIndex)
	var dir = dirEntry{abs: abs, providedPath: providedPath}
	for i, dirEntry := range entries {
		dir.dirEntry = dirEntry
		t.dirEntries[index+i] = dir
	}

	return
}

// directoryOrder returns unsorted entries, most with deferred [fs.FileInfo]
func (t *Traverser) directoryOrder(abs string) (entries []fs.DirEntry, err error) {
	// open the directory
	var osFile *os.File
	if osFile, err = os.Open(abs); perrors.Is(&err, "os.Open %w", err) {
		return
	}
	defer parl.Close(osFile, &err)

	// reads in directory order
	//	- returns an array of interface-value pointers, meaning
	//		each DirEntry is a separate allocation
	//	- for most modeTypes, Type is availble without lstat invocation
	//	- lstat is then deferred until Info method is invoked.
	//		Every time such deferring Info is invoked, lstat is executed
	if entries, err = osFile.ReadDir(allNamesAtOnce); perrors.Is(&err, "os.File.ReadDir %w", err) {
		return
	}

	return
}
