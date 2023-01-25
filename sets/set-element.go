/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

type SetElement[T comparable] struct {
	ValueV T
	Name   string
}

var _ Element[int] = &SetElement[int]{}

func (pv *SetElement[T]) Value() (value T) {
	return pv.ValueV
}

func (pv *SetElement[T]) String() (s string) {
	return pv.Name
}
