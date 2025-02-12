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
	// key that maps to this enumeration value
	KeyK K
	// sentence describing this flag
	Full string
}

var _ sets.Element[int] = &EnumItemImpl[int, int]{}

// NewEnumItemImpl returns a concrete value representing an element of an ordered collection
// used to implement Integer enumeratiopns and bit-fields
func NewEnumItemImpl[K constraints.Ordered, V any](value V, key K, full string) (item parl.EnumItem[K, V]) {
	return &EnumItemImpl[K, V]{KeyK: key, Full: full, ValueV: value}
}

// Description returns a descriptive sentence for this enumeration value
func (item *EnumItemImpl[K, V]) Description() (desc string) { return item.Full }

// Key returns the key for this enumeration value
func (item *EnumItemImpl[K, V]) Key() (key K) { return item.KeyK }

// Value returns this enumeration value’s value using the restricted type
func (item *EnumItemImpl[K, V]) Value() (value V) { return item.ValueV }

func (item *EnumItemImpl[K, V]) String() (s string) {
	return parl.Sprintf("key:%v'%s'0x%x", item.KeyK, item.Full, item.ValueV)
}
