/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"golang.org/x/exp/constraints"
)

// AssignedPriority contains the assigned priority for a priority-queue element
//   - V is the element value type whose pointer-value provides identity
//   - P is the priority, a descending-ordered type
//   - SetPriority updates the priority
//   - Cmp makes AssignedPriority ordered
type AssignedPriority[V any, P constraints.Ordered] struct {
	Priority P   // the main sort value, ordered high to low
	Index    int // insertion order: lowest/earliest value first: distinguishes between equal priorities
	Value    *V  // the pointer provides identity for this priority
}

func NewAssignedPriority[V any, P constraints.Ordered](priority P, index int, value *V) (assignedPriority *AssignedPriority[V, P]) {
	return &AssignedPriority[V, P]{Priority: priority, Index: index, Value: value}
}

func (ap *AssignedPriority[V, P]) SetPriority(priority P) {
	ap.Priority = priority
}

// Cmp sorts descending: -1 results appears first
func (a *AssignedPriority[V, P]) Cmp(b *AssignedPriority[V, P]) (result int) {
	if a.Priority > b.Priority {
		return -1
	} else if a.Priority < b.Priority {
		return 1
	} else if a.Index < b.Index {
		return -1
	} else if a.Index > b.Index {
		return 1
	}
	return 0
}
