/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"fmt"
	"os"

	"github.com/google/btree"
)

// BtreeIterator retrieves B-tree values in order
type BtreeIterator[V any] struct {
	tree      *btree.BTreeG[V]
	list      []V
	listIndex int
}

// NewBtreeIterator returns an object that can retrieve elements in order
func NewBtreeIterator[V any](tree *btree.BTreeG[V]) (iterator *BtreeIterator[V]) {
	return &BtreeIterator[V]{tree: tree}
}

// func(item T) bool
var _ btree.ItemIteratorG[int]

// Iterate returns a sorted list of the first n elements
func (b *BtreeIterator[V]) Iterate(n int) (list []V) {

	// create list
	list = make([]V, n)
	b.list = list

	// traverse
	b.listIndex = 0
	b.tree.Ascend(b.iterator)

	return
}

// iterator receives elements from the B-tree
func (b *BtreeIterator[V]) iterator(value V) (keepGoing bool) {
	if b.listIndex >= len(b.list) {
		fmt.Fprintf(os.Stderr, "pmaps.BtreeIterator[…].iterator %d >= %d", b.listIndex, len(b.list))
		return
	}

	b.list[b.listIndex] = value
	b.listIndex++

	keepGoing = b.listIndex < len(b.list)
	return
}
