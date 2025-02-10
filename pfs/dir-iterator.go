/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/iters"
)

// DirIterator traverses file-system directories
type DirIterator struct {
	// traverser is the file-system traverser.
	// Only method is [Traverser.Next]
	traverser Traverser
	// BaseIterator provides iterator methods:
	// [iters.Iterator.Cancel] [iters.Iterator.Cond] [iters.Iterator.Next]
	iters.BaseIterator[ResultEntry]
}

// NewDirIterator returns an iterator for directories
//   - path is the initial path for the file-system walk.
//     it may be relative or absolute, contain symlinks and
//     point to a file, directory or special file
//   - only directories are returned.
//     if directories are not skipped, they descended into.
//   - symlinks are followed.
//     Broken symlinks are ignored.
//   - any errored file-system entry cancels the iterator with error.
func NewDirIterator(path string) (iterator iters.Iterator[ResultEntry]) {
	i := DirIterator{traverser: *NewTraverser(path)}
	i.BaseIterator = *iters.NewBaseIterator(i.iteratorAction)
	return &i
}

// Init implements the right-hand side of a short variable declaration in
// the init statement of a Go “for” clause
//
//		for i, iterator := iters.NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
//	   // i is pointer to slice element
func (i *DirIterator) Init() (result ResultEntry, iterator iters.Iterator[ResultEntry]) {
	iterator = i
	return
}

// iteratorAction provides items to the BaseIterator
func (t *DirIterator) iteratorAction(isCancel bool) (result ResultEntry, err error) {
	if isCancel {
		return
	}
	for {
		result = t.traverser.Next()

		// handle end
		if result.IsEnd() {
			err = parl.ErrEndCallbacks
			return
		}

		// ignore broken symlink
		if result.Reason == RSymlinkBad {
			continue
		}

		// any other error cancels iterator
		if result.Err != nil {
			err = result.Err
			return
		}

		//	- ignore any file-system entry that is not
		//		a directory
		if !result.IsDir() {
			continue
		}

		return // directory return
	}
}
