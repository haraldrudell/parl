/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Ranking is a pointer-identity-to-value map of updatable values traversable by rank.
// Ranking implements [parl.Ranking][V comparable, R constraints.Ordered].
package pmaps

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices"
	"golang.org/x/exp/constraints"
)

// PriorityQueue is a pointer-identity-to-value map of updatable values traversable by rank.
// PriorityQueue implements [parl.PriorityQueue][V comparable, R constraints.Ordered].
//   - V is a value reference composite type that is comparable, ie. not slice map function.
//     Preferrably, V is interface or pointer to struct type.
//   - R is an ordered type such as int floating-point string, used to rank the V values
//   - values are added or updated using AddOrUpdate method distinguished by
//     (computer science) identity
//   - if the same comparable value V is added again, that value is re-ranked
//   - rank R is computed from a value V using the ranker function.
//     The ranker function may be examining field values of a struct
//   - values can have the same rank. If they do, equal rank is provided in insertion order
type PriorityQueue[V any, P constraints.Ordered] struct {
	// priorityFunc is the function computing priority for a value-pointer
	priorityFunc func(value *V) (priority P)
	// queue is a list of queue nodes ordered by descending priority
	queue parl.Ordered[*AssignedPriority[V, P]]
	// m is a map providing O(1) access to ranking nodes by value-pointer
	m map[*V]*AssignedPriority[V, P]
}

// NewPriorityQueue returns a map of updatable values traversable by rank
func NewPriorityQueue[V any, P constraints.Ordered](
	priorityFunc func(value *V) (priority P),
) (priorityQueue parl.PriorityQueue[V, P]) {
	if priorityFunc == nil {
		perrors.NewPF("ranker cannot be nil")
	}
	r := PriorityQueue[V, P]{
		priorityFunc: priorityFunc,
		m:            map[*V]*AssignedPriority[V, P]{},
	}
	r.queue = pslices.NewOrderedAny(r.comparatorFunc)
	return &r
}

// AddOrUpdate adds a new value to the ranking or updates the ranking of a value
// that has changed.
func (pq *PriorityQueue[V, P]) AddOrUpdate(valuep *V) {

	// obtain updated priority
	priority := pq.priorityFunc(valuep)

	var assignedPriority *AssignedPriority[V, P]
	var ok bool
	if assignedPriority, ok = pq.m[valuep]; !ok {

		// new value case
		assignedPriority = NewAssignedPriority(priority, pq.queue.Length(), valuep)
		pq.m[valuep] = assignedPriority
	} else {

		// node update case
		pq.queue.Delete(assignedPriority)
		assignedPriority.SetPriority(priority)
	}

	// update order: that’s what we do here. We keep order
	pq.queue.Insert(assignedPriority)
}

// List returns the first n or default all values by rank
func (pq *PriorityQueue[V, P]) List(n ...int) (valueQueue []*V) {

	// get number of items n0
	var n0 int
	length := pq.queue.Length()
	if len(n) > 0 {
		n0 = n[0]
	}
	if n0 < 1 || n0 > length {
		n0 = length
	}

	// build list
	valueQueue = make([]*V, n0)
	assignedPriorityQueue := pq.queue.List()
	for i := 0; i < n0; i++ {
		valueQueue[i] = assignedPriorityQueue[i].Value
	}
	return
}

// comparatorFunc is provided to pslices.NewOrderedAny based on AssignedPriority.Cmp
func (pq *PriorityQueue[V, P]) comparatorFunc(a, b *AssignedPriority[V, P]) (result int) {
	return a.Cmp(b)
}
