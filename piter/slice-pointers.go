/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package piter

import "iter"

type SlicePointerIterator[T any] []T

// SlicePointers().R is iterator over pointers to slice elements and the slice indices for those elements
var _ iter.Seq2[*string, int] = SlicePointers([]string{}).R

func SlicePointers[T any](slice []T) (p SlicePointerIterator[T]) {
	return SlicePointerIterator[T](slice)
}

func (p SlicePointerIterator[T]) R(yield func(tp *T, index int) (keepGoing bool)) {
	for i := range len(p) {
		if !yield(&p[i], i) {
			return
		}
	}
}

func (p SlicePointerIterator[T]) Reverse(yield func(tp *T, index int) (keepGoing bool)) {
	for i := len(p) - 1; i >= 0; i-- {
		if !yield(&p[i], i) {
			return
		}
	}
}
