/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"cmp"

	"github.com/google/btree"
)

// CmpLess provides a Less function based on a comparison function
//   - Less returns true when a < b
//   - Less returns false for equal or b > a
//   - the Less function is defined by github.com/google/btree and cmp packages
//   - the compare function is defined by the cmp package
//   - CmpLess is required since a constructor cannot store references
//     to its own fields in order to avoid memory leaks and corrupt behaviors.
//     The constructor can, however, store pointers to a CmpLess object
type CmpLess[V any] struct {
	cmp func(a, b V) (result int)
}

// NewCmpLess returns a Less function object from a comparison function
func NewCmpLess[V any](cmp func(a, b V) (result int)) (cmpLess *CmpLess[V]) {
	return &CmpLess[V]{cmp: cmp}
}

// func(a, b T) bool
var _ btree.LessFunc[int]

// Less reports whether x is less than y
//   - false for equal
var _ = cmp.Less[int]

// Compare returns
//   - -1 if x is less than y,
//   - 0 if x equals y,
//   - +1 if x is greater than y.
var _ = cmp.Compare[int]

// Less is a Less function based on a compare function
func (c *CmpLess[V]) Less(a, b V) (aIsFirst bool) {
	return c.cmp(a, b) < 0
}
