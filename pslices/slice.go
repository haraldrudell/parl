/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import "golang.org/x/exp/slices"

// Slice implements basic slice methods for use by a slice embedded in a struct.
// Slice implements parl.Slice[E any].
type Slice[E any] struct {
	list []E
}

func (o *Slice[E]) Element(index int) (element E) {
	if index >= 0 && index < len(o.list) {
		element = o.list[index]
	}
	return
}

func (o *Slice[E]) Length() (index int) {
	return len(o.list)
}

func (o *Slice[E]) Clear() {
	o.list = o.list[:0]
}

// List returns a clone of the ordered slice.
func (o *Slice[E]) List() (list []E) {
	return slices.Clone(o.list)
}
