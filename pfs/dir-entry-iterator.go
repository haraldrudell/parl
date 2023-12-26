/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"io/fs"
	"path/filepath"
	"sync"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/iters"
)

type DirEntry struct {
	fs.DirEntry
	ProvidedPath string
}

type DirEntryIterator struct {
	path string
	iters.BaseIterator[DirEntry]
	sliceOnce sync.Once
	err       error

	entriesLock sync.Mutex
	entries     []fs.DirEntry
}

func NewDirEntryIterator(path string) (iterator iters.Iterator[DirEntry]) {
	i := DirEntryIterator{path: path}
	i.BaseIterator = *iters.NewBaseIterator(i.iteratorAction)
	return &i
}

// Init implements the right-hand side of a short variable declaration in
// the init statement of a Go “for” clause
//
//		for i, iterator := iters.NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
//	   // i is pointer to slice element
func (i *DirEntryIterator) Init() (result DirEntry, iterator iters.Iterator[DirEntry]) {
	iterator = i
	return
}

// iteratorAction provides items to the BaseIterator
func (t *DirEntryIterator) iteratorAction(isCancel bool) (result DirEntry, err error) {
	if isCancel {
		return
	}
	t.sliceOnce.Do(t.loadSlice)
	if err = t.err; err != nil {
		return
	}
	result.DirEntry = t.entry()
	if result.DirEntry == nil {
		err = parl.ErrEndCallbacks
		return
	}
	result.ProvidedPath = filepath.Join(t.path, result.Name())
	return
}

func (t *DirEntryIterator) entry() (entry fs.DirEntry) {
	t.entriesLock.Lock()
	defer t.entriesLock.Unlock()

	if len(t.entries) == 0 {
		return
	}
	entry = t.entries[0]
	t.entries[0] = nil
	t.entries = t.entries[1:]

	return
}
func (t *DirEntryIterator) loadSlice() { t.entries, t.err = ReadDir(t.path) }
