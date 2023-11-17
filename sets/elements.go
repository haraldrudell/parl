/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import (
	"github.com/haraldrudell/parl/iters"
)

// NewElements returns an iterator of interface-type sets.Element[T]
//   - elements is a slice of a concrete type, named E, that should implement
//     sets.Element
//   - at compile time, elements is slice of any: []any
//   - based on a slice of non-interface-type Elements[T comparable].
func NewElements[T comparable, E any](elements []E) (iter iters.Iterator[Element[T]]) {
	return iters.NewSliceInterfaceIterator[Element[T]](elements)
}
