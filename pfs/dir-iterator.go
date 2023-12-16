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

// NewIterator returns an iterator for a file-system entry and any child entries if directory
//   - path is the initial path for the file-system walk.
//     it may be relative or absolute, contain symlinks and
//     point to a file, directory or special file
//   - only error-free directories are returned.
//     if directories are not skipped, they are followed.
//   - any file-system entry error cancels the iterator with error.
//     Entry end return ends the iterator.
//   - symlinks are followed
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

		// any error cancels iterator
		if result.Err != nil {
			err = result.Err
			return
		}

		//	- if it is not an error or end,
		//	- and it is not a directory,
		//	- ignore it
		if !result.IsDir() {
			continue
		}

		return // directory return
	}
}
