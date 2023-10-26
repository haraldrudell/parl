/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package enum

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/sets"
	"golang.org/x/exp/constraints"
)

// EnumItemImpl is an value or flag-value item of an enumeration container.
//   - K is the type of a unique key mapping one-to-one to an enumeration value
//   - V is the type of the internally used value representation
type EnumItemImpl[K constraints.Ordered, V any] struct {
	ValueV V
	KeyK   K      // key that maps to this enumeration value
	Full   string // sentence describing this flag
}

var _ sets.Element[int] = &EnumItemImpl[int, int]{}

func NewEnumItemImpl[K constraints.Ordered, V any](value V, key K, full string) (item parl.EnumItem[K, V]) {
	return &EnumItemImpl[K, V]{KeyK: key, Full: full, ValueV: value}
}

func (item *EnumItemImpl[K, V]) Description() (desc string) {
	return item.Full
}

func (item *EnumItemImpl[K, V]) Key() (key K) {
	return item.KeyK
}

func (item *EnumItemImpl[K, V]) Value() (value V) {
	return item.ValueV
}

func (item *EnumItemImpl[K, V]) String() (s string) {
	return parl.Sprintf("key:%v'%s'0x%x", item.KeyK, item.Full, item.ValueV)
}
