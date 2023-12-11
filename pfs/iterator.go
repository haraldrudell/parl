/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import "github.com/haraldrudell/parl/iters"

type Iterator struct {
	traverser Traverser
	iters.BaseIterator[ResultEntry]
}

//   - rootPath is the initial path for the file-system walk.
//     it may be relative or absolute, contain symlinks and
//     point to a file, directory or special file
func NewIterator(path string) (iterator iters.Iterator[ResultEntry]) {
	i := Iterator{traverser: *NewTraverser(path)}
	i.BaseIterator = *iters.NewBaseIterator(i.iteratorAction)
	return &i
}

func (i *Iterator) Init() (result ResultEntry, iterator iters.Iterator[ResultEntry]) {
	iterator = i
	return
}

func (t *Iterator) iteratorAction(isCancel bool) (result ResultEntry, err error) {
	if isCancel {
		return
	}
	result = t.traverser.Next()

	return
}
