/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

// SetElementFull is an value or flag-value item of an enumeration container.
//   - K is the type of a unique key mapping one-to-one to an enumeration value
//   - V is the type of the internally used value representation
type SetElementFull[T comparable] struct {
	ValueV T
	Name   string // key that maps to this enumeration value
	Full   string // sentence describing this flag
}

func (item *SetElementFull[T]) Description() (desc string) {
	return item.Full
}

func (item *SetElementFull[T]) Value() (value T) {
	return item.ValueV
}

func (item *SetElementFull[T]) String() (s string) {
	return item.Name
}
