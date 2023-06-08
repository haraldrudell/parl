/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"golang.org/x/exp/slices"
)

// Slice is a base type for reusable slice implementations
//   - super-types could implement Insert Delete Index
//   - Length
//   - Element single-element access
//   - Slice multiple-element access
//   - DeleteIndex delete elements
//   - Clear Clone
//   - List clone of n first items
//   - Slice implements parl.Slice[E any].
type Slice[E any] struct {
	list []E
}

// Element returns element by index.
// if index is negative or the length of the slice or larger, the E zero-value is returned.
func (o *Slice[E]) Element(index int) (element E) {
	if index >= 0 && index < len(o.list) {
		element = o.list[index]
	}
	return
}

// Slice returns a multiple-element sub-slice
func (o *Slice[E]) SubSlice(index0, index1 int) (elements []E) {
	if index0 >= 0 && index0 <= index1 && index1 <= len(o.list) {
		elements = o.list[index0:index1]
	}
	return
}

func (o *Slice[E]) SetElement(index int, element E) {
	if index < 0 {
		index = 0
	}
	if index >= cap(o.list) {
		list := o.list
		o.list = make([]E, index+1)
		copy(o.list, list)
	} else if index >= len(o.list) {
		o.list = o.list[:index+1]
	}
	o.list[index] = element
}

// Append adds element at end
func (o *Slice[E]) Append(slice []E) {
	o.list = append(o.list, slice...)
}

// DeleteIndex removes elements by index
//   - index1 default is slice length
func (o *Slice[E]) DeleteIndex(index0 int, index1 ...int) {
	length := len(o.list)
	var index2 int
	if len(index1) > 0 {
		index2 = index1[0]
	} else {
		index2 = length
	}

	// ensure indexes are valid
	if index0 < 0 {
		index0 = 0
	} else if index0 > length {
		index0 = length
	}
	if index2 < index0 {
		index2 = index0
	} else if index2 > length {
		index2 = length
	}

	// execute delete
	if index0 == index2 {
		return // nothing to do return
	} else if index0 == 0 {

		// delete at beginning
		TrimLeft(&o.list, index2)
		return
	} else if index2 == length {

		// delete at end
		SetLength(&o.list, index0)
		return // delete at end return
	}

	// delete in the middle
	deleteCount := index2 - index0
	copy(o.list[index0:], o.list[index2:])
	SetLength(&o.list, length-deleteCount)
}

// Length returns number of elements in the slice
func (o *Slice[E]) Length() (index int) {
	return len(o.list)
}

// Cap returns slice capacity
func (o *Slice[E]) Cap() (capacity int) {
	return cap(o.list)
}

// Clear empties the ordered slice
func (o *Slice[E]) Clear() {
	o.list = o.list[:0]
}

// Clone returns a shallow clone of the slice
func (o *Slice[E]) Clone() (clone *Slice[E]) {
	return &Slice[E]{list: slices.Clone(o.list)}
}

// List returns a clone of the n or all first slice elements
func (o *Slice[E]) List(n ...int) (list []E) {
	length := o.Length()

	// get number of items n0
	var n0 int
	if len(n) > 0 {
		n0 = n[0]
	}
	if n0 < 1 || n0 > length {
		n0 = length
	}

	return slices.Clone(o.list[:n0])
}
