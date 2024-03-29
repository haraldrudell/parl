/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omaps

import (
	"github.com/google/btree"
	"golang.org/x/exp/constraints"
)

// LessOrdered is [btree.LessFunc] for ordered values
func LessOrdered[V constraints.Ordered](a, b V) (aBeforeB bool) { return a < b }

var _ btree.LessFunc[int] = LessOrdered[int]
