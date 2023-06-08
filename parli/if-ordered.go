/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parli

// OrderedPointers[E any] is an ordered list of pointers sorted by the referenced values.
// OrderedPointers should be used when E is a large struct.
// pslices.NewOrderedPointers[E constraints.Ordered]() instantiates for pointers to
// comparable types, ie. not func slice map.
// pslices.NewOrderedAny instantiates for pointers to any type or for custom sort orders.
type OrderedPointers[E any] interface {
	Ordered[*E]
}

// Ordered[E any] is an ordered list of values.
//   - Ordered should be used when E is interface type or a small-sized value.
//   - pslices.NewOrdered[E constraints.Ordered]() instantiates for
//     comparable types, ie. not func map slice
//   - pslices.NewOrderedAny[E any](cmp func(a, b E) (result int)) instantiates for any type
//     or for custom sort order
//   - pslices.NewOrderedPointers[E constraints.Ordered]() orders pointer to value, to
//     be used for large structs or order of in-place data
//   - those list implementations have Index O(log n)
type Ordered[E any] interface {
	Slice[E] // Element() Length() Clear() List()
	// Insert adds an element to the ordered slice.
	// The implementation may allow duplicates.
	Insert(element E)
	// Delete removes an element from the ordered slice.
	// If the ordered slice does not contain element, the slice is not changed.
	Delete(element E)
	// Index returns index of the first occurrence of element in the ordered slice
	// or -1 if element is not in the slice.
	Index(element E) (index int)
	// Clone returns a clone of the ordered slice
	Clone() (ordered Ordered[E])
}

// Slice is a reusable unordered slice
//   - functions by index that apply to unordered slice
//   - embedded interface for slice interface types
type Slice[E any] interface {
	// Element returns element by index.
	// if index is negative or the length of the slice or larger, the E zero-value is returned.
	Element(index int) (element E)
	// SubSlice returns a multiple-element sub-slice, not a clone
	//	index0 and index1 aree adjusted to legal values
	SubSlice(index0, index1 int) (elements []E)
	// SetElement updates single element by index
	//	- slice is extended if too short
	SetElement(index int, element E)
	// Append adds element at end
	Append(slice []E)
	// DeleteIndex removes elements by index
	//	- index1 default is slice length
	DeleteIndex(index0 int, index1 ...int)
	// Length returns number of elements in the slice
	Length() (length int)
	// Cap returns slice capacity
	Cap() (capacity int)
	// Clear empties the ordered slice
	Clear()
	// List returns a clone of the internal slice
	List(n ...int) (list []E)
}
