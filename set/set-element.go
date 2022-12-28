/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package set

type Element[T comparable] struct {
	ValueV T
	Name   string
}

func (pv *Element[T]) Value() (value T) {
	return pv.ValueV
}

func (pv *Element[T]) String() (s string) {
	return pv.Name
}
