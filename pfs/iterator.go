/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/iters"
)

// Iterator traverses file-system entries and errors
type Iterator struct {
	// traverser is the file-system traverser.
	// Only method is [Traverser.Next]
	traverser Traverser
	// BaseIterator provides iterator methods:
	// [iters.Iterator.Cancel] [iters.Iterator.Cond] [iters.Iterator.Next]
	iters.BaseIterator[ResultEntry]
}

// NewIterator returns a directory-tree iterator for all file-system entries following symbolic links
//   - path: initial path for file-system traversal.
//     path may be relative or absolute, contain symlinks and
//     point to a file, directory or special file
//   - entries that are error-free non-directory non-symbolic-link are returned as [REntry]
//   - symbolic links and directories that are not actively skipped are followed
//   - — returned as [RSkippable]
//   - “.” and “..” are not returned
//   - directories that cannot be opened or listed are returned again as [RDirBad]
//   - broken symbolic links are returned as [RSymlinkBad]
//   - entries that unexpectedly produce error are returned as [RError]
//   - —
//   - [NewDirEntryIterator] single-level directory iterator
//   - delegates to [pfs.NewTraverser]
func NewIterator(path string) (iterator iters.Iterator[ResultEntry]) {
	i := Iterator{}
	NewTraverser(path, &i.traverser)
	i.BaseIterator = *iters.NewBaseIterator(i.iteratorAction)
	return &i
}

// Init implements the right-hand side of a short variable declaration in
// the init statement of a Go “for” clause
//
//		for i, iterator := iters.NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
//	   // i is pointer to slice element
func (i *Iterator) Init() (result ResultEntry, iterator iters.Iterator[ResultEntry]) {
	iterator = i
	return
}

// iteratorAction provides items to the BaseIterator
func (t *Iterator) iteratorAction(isCancel bool) (result ResultEntry, err error) {
	if isCancel {
		return // cancel notify return: Tarverser has no cleanup
	}

	// get next file-system entry or error
	result = t.traverser.Next()

	//end iterator when traverser ends
	if result.IsEnd() {
		err = parl.ErrEndCallbacks
	}

	return
}
