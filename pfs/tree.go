/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"errors"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/haraldrudell/parl"
)

const (
	sSep = string(filepath.Separator)
)

// Tree represents a file system scan originating at a single absolute or relative path
//   - additional roots may appear if the original path contains symlinks
//     that point outside the original directory tree
//   - each such root is a separate starting path in the overall file system
type Tree struct {
	// map of the absolute, clean paths for each encountered root
	//	- key: absolute, clean file-system path to root file or directory
	//	- value: index in roots
	//	- because roots slice may be re-allocated, value index an not root pointer
	rootsRegistry PathRegistry[Root]
	// obsoleteRoots will be encountered during scanning of
	// symbolic links targeting above the root in the file system
	obsoleteRoots PathRegistry[Root]
	// a collection of encountered symbolic link targets that
	// may create additional roots
	//	- path value has no symlinks and are clean and absolute
	symlinkRegistry PathRegistry[string]
	walkFunc        filepath.WalkFunc
	FSEntryCount    *atomic.Uint64
}

// NewTree returns a file-system scan object
func NewTree(walkFunc filepath.WalkFunc) (tree *Tree) {
	var u64 atomic.Uint64
	return &Tree{
		rootsRegistry:   *NewPathRegistry[Root](),
		obsoleteRoots:   *NewPathRegistry[Root](),
		symlinkRegistry: *NewPathRegistry[string](),
		walkFunc:        walkFunc,
		FSEntryCount:    &u64,
	}
}

func (t *Tree) SetCounter(FSEntryCount *atomic.Uint64) {
	t.FSEntryCount = FSEntryCount
}

// ScanRoots scans rootpath and all additional encountered roots
//   - rootPath is as provided to the Walk function, ie. may be relative
//     and unclean
func (t *Tree) ScanRoots(rootPath string) (err error) {

	// scan the initial root
	if err = t.scanRoot(rootPath); err != nil {
		return
	}

	// process all symlinks
	for {
		var symlinkTarget string
		if sp := t.symlinkRegistry.GetNext(); sp != nil {
			symlinkTarget = *sp
		} else {
			break
		}

		t.processSymlink(symlinkTarget)
	}

	// drop registries
	t.obsoleteRoots.Drop()
	t.symlinkRegistry.Drop()
	t.rootsRegistry.Drop(DropPathsNotValues)

	return
}

// Walk traverses the built tree
func (t *Tree) Walk() (err error) {
	if parl.IsThisDebug() {
		St("Tree.Walk roots: %d", t.rootsRegistry.MapLength())
	}
	for {
		var root = t.rootsRegistry.GetNext() // GetNext slices away until values is empty
		if root == nil {
			return
		}
		if err = t.walkRoot(NewPendingEntry("", root.ProvidedPath, root.FSEntry)); err != nil {
			if errors.Is(err, filepath.SkipAll) {
				err = nil // skipAll: return without error
			}
			return // error or skipAll
		}
	}
}

func (t *Tree) walkRoot(pendings *PendingEntry) (err error) {
	var lastPending = pendings
	for pendings != nil {

		// get data for pending entry
		var path = pendings.Path
		var entry = pendings.Entry
		if pendings = pendings.Next; pendings == nil {
			lastPending = nil
		}
		var info, e = entry.Walks()

		// walkFunc and its errors
		if err = t.walkFunc(path, info, e); err != nil {
			if entry.IsDir() && errors.Is(err, filepath.SkipDir) {
				err = nil
				continue // skipDir: continue with next
			}
			return // error return
		}

		// nothing more to do for files
		if !entry.IsDir() {
			continue
		}

		var dirEntry = entry.(*Directory)
		for _, child := range dirEntry.Children {
			var p = filepath.Join(path, child.Name())
			if child.IsDir() {
				var pending = NewPendingEntry("", p, child)
				if pendings == nil {
					pendings = pending
				} else {
					lastPending.Next = pending
				}
				lastPending = pending
				continue
			}

			// child is file
			info, e = child.Walks()
			if err = t.walkFunc(p, info, e); err != nil {
				return
			}
		}
	}
	return
}

func (t *Tree) scanRoot(rootPath string) (err error) {

	// create a root for rootPath
	var root = NewRoot(rootPath)
	var absPath string
	var rootEntry FSEntry
	if absPath, rootEntry, err = root.Init(); err != nil {
		return
	}
	t.FSEntryCount.Add(1)

	if parl.IsThisDebug() {
		St("scanRoot: %q abs: %q", rootPath, absPath)
	}

	// save the new root in root registry
	t.rootsRegistry.Add(absPath, root)

	// scan the root collecting its symlinks
	err = NewEntryScanner(rootEntry, absPath, rootPath, t.addSymlink, t.FSEntryCount).Scan()

	return
}

// addSymlink stores an encountered symlink for later processing
func (t *Tree) addSymlink(abs string) {
	t.symlinkRegistry.Add(abs, &abs)
}

// processSymlink patches the tree for a symlink
func (t *Tree) processSymlink(absTarget string) (symlinkTargets []string, err error) {

	// check for exact match to existing root
	if t.rootsRegistry.HasAbs(absTarget) {
		return // symlink matches existing root: ignore it
	}

	// match absTarget against existing roots
	var length = t.rootsRegistry.ListLength()
	for i := 0; i < length; {
		var root = t.rootsRegistry.GetValue(i)
		if root == nil {
			continue // a discarded value
		}
		var rootAbs = root.Abs()

		// if absTarget is a subdirectory of an existing root, it can be ignored
		if strings.HasPrefix(absTarget+sSep, rootAbs+sSep) {
			return // symlink is a sub-entry of an existing root: ignore it
		}

		// if root is not a subdirectory of absTarget, check the next root
		if !strings.HasPrefix(rootAbs+sSep, absTarget+sSep) {
			i++
			continue
		}

		// root is a subdirectory of this symlink, obsolete the root
		t.obsoleteRoots.Add(rootAbs, root)
		t.rootsRegistry.DeleteByIndex(i)
		break // scan the symlink as a new root
	}

	// scan as new root
	//	- a new, separate root or
	//	- a root is a subdirectory of this symlink
	err = t.scanRoot(absTarget)

	return
}
