/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omaps

import (
	"fmt"
	"os"

	"github.com/google/btree"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices"
)

// BtreeIteratorFunc is a function that converts B-tree V-values
// to slice storage T values
//   - valueToStore nil: ignore the value
type BtreeIteratorFunc[V any, T any] func(value V) (valueToStore T, err error)

// BtreeIterator retrieves B-tree values in order
//   - V the type of values tree holds
//   - T the type of values in the returned list
type BtreeIterator[V any, T any] struct {
	tree      *btree.BTreeG[V]        // the tree to iterate over
	converter BtreeIteratorFunc[V, T] // optional converter function
	// the list where to store values
	//	- slice length is how many values should be returned
	list      []T
	listIndex int // next storage position in list
	err       error
}

// NewBtreeIterator returns an object that can retrieve elements in order
//   - V the type of values tree holds
//   - T the type of values in the returned list
func NewBtreeIterator[V any, T any](
	tree *btree.BTreeG[V], // the tree to iterate over
	converter ...BtreeIteratorFunc[V, T], // optionl converter function
) (iterator *BtreeIterator[V, T]) {
	var c BtreeIteratorFunc[V, T]
	if len(converter) > 0 {
		c = converter[0]
	}
	return &BtreeIterator[V, T]{
		tree:      tree,
		converter: c,
	}
}

// func(item T) bool
var _ btree.ItemIteratorG[int]

// Iterate returns a sorted list of the first n elements
//   - n is how many items will be returned
//   - list is an optional list where to store results
//   - results is populated list
func (i *BtreeIterator[V, T]) Iterate(n int, list ...[]T) (results []T, err error) {
	if n == 0 {
		return // n 0 return: results nil
	}
	i.list = nil
	i.listIndex = 0

	// ensure results list of length n
	if len(list) > 0 {
		i.list = list[0] // use provided list
	}
	if i.list == nil {
		i.list = make([]T, n) // create slice of correct length
	} else if len(i.list) < n {
		pslices.SetLength(&i.list, n) // ensure slice of correct capacity and length
	}

	// traverse
	i.tree.Ascend(i.iterator)
	results = i.list
	i.list = nil

	return
}

// iterator receives elements of type V from the B-tree
func (i *BtreeIterator[V, T]) iterator(value V) (keepGoing bool) {
	if i.listIndex >= len(i.list) {
		fmt.Fprintf(os.Stderr, "pmaps.BtreeIterator[…].iterator %d >= %d", i.listIndex, len(i.list))
		return // invoked after false returned error
	}

	// convert V to T
	var result T
	if f := i.converter; f != nil {
		result, i.err = f(value)
	} else {
		var v any = value
		var ok bool
		if result, ok = v.(T); !ok {
			i.err = perrors.ErrorfPF("cannot convert %T to %T", value, result)
			return // bad conversion return: keepGoin false
		}
	}

	i.list[i.listIndex] = result
	i.listIndex++
	keepGoing = i.listIndex < len(i.list)

	return
}
