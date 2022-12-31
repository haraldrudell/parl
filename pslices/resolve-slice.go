/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// ResolveSlice removes one level of indirection from a slice of pointers.
package pslices

// ResolveSlice removes one level of indirection from a slice of pointers.
func ResolveSlice[E any](slic []*E) (sList []E) {
	length := len(slic)
	if length == 0 {
		return
	}
	sList = make([]E, length)
	for i, e := range slic {
		sList[i] = *e
	}
	return
}
