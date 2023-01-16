/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Ranking is a pointer-identity-to-value map of updatable values traversable by rank.
// Ranking implements [parli.Ranking][V comparable, R constraints.Ordered].
package pmaps

import (
	"github.com/haraldrudell/parl/ids"
	"github.com/haraldrudell/parl/parli"
	"github.com/haraldrudell/parl/pslices"
	"golang.org/x/exp/constraints"
)

type AggregatingPriorityQueue[V any, P constraints.Ordered] struct {
	// queue is a list of queue nodes ordered by rank
	queue parli.Ordered[parli.AggregatePriority[V, P]]
	// m provides O(1) access to priority data-nodes via the value-pointer
	m map[*V]parli.AggregatePriority[V, P]
	// indexGenerator provides IDs for insertion-ordering
	indexGenerator ids.UniqueIDint
}

// NewRanking returns a map of updatable values traversable by rank
func NewAggregatingPriorityQueue[V any, P constraints.Ordered]() (priorityQueue parli.AggregatingPriorityQueue[V, P]) {
	p := AggregatingPriorityQueue[V, P]{
		m: map[*V]parli.AggregatePriority[V, P]{},
	}
	p.queue = pslices.NewOrderedAny(p.Cmp)
	return &p
}

// Get retrieves a possible value container associated with valuep
func (pq *AggregatingPriorityQueue[V, P]) Get(valuep *V) (aggregator parli.Aggregator[V, P], ok bool) {
	var nodep parli.AggregatePriority[V, P]
	if nodep, ok = pq.m[valuep]; ok {
		aggregator = nodep.Aggregator()
	}
	return
}

// Put stores a new value container associated with valuep
//   - the valuep is asusmed to not have a node in the queue
func (pq *AggregatingPriorityQueue[V, P]) Put(valuep *V, aggregator parli.Aggregator[V, P]) {

	// create aggregatePriority with current priority from aggregator
	aggregatePriority := NewAggregatePriority(
		valuep,
		pq.indexGenerator.ID(),
		aggregator,
	)

	// store in map
	pq.m[valuep] = aggregatePriority

	// update order: that’s what we do here. We keep order
	pq.queue.Insert(aggregatePriority)
}

// Update re-prioritizes a value
func (pq *AggregatingPriorityQueue[V, P]) Update(valuep *V) {

	var aggregatePriority parli.AggregatePriority[V, P]
	var ok bool
	if aggregatePriority, ok = pq.m[valuep]; !ok {
		return // value priority does not exist return
	}

	// update order: that’s what we do here. We keep order
	pq.queue.Delete(aggregatePriority)
	aggregatePriority.Update()
	pq.queue.Insert(aggregatePriority)
}

// Clear empties the priority queue. The hashmap is left intact.
func (pq *AggregatingPriorityQueue[V, P]) Clear() {
	pq.queue.Clear()
}

// List returns the first n or default all values by pirority
func (pq *AggregatingPriorityQueue[V, P]) List(n ...int) (aggregatorQueue []parli.AggregatePriority[V, P]) {
	return pq.queue.List(n...)
}

// Cmp returns a comparison of two AggregatePriority objects that represents value elements.
//   - Cmp is a custom comparison function to be sued with pslices and slices packages
//   - Cmp makes AggregatePriority ordered
func (pq *AggregatingPriorityQueue[V, P]) Cmp(a, b parli.AggregatePriority[V, P]) (result int) {
	aPriority := a.Aggregator().Priority()
	bPriority := b.Aggregator().Priority()
	if aPriority > bPriority { // highest priority first
		return -1
	} else if aPriority < bPriority {
		return 1
	}
	aIndex := a.(*AggregatePriority[V, P]).assignedPriority.Index
	bIndex := b.(*AggregatePriority[V, P]).assignedPriority.Index
	if aIndex < bIndex { // lowest index first
		return -1
	} else if aIndex > bIndex {
		return 1
	}
	return 0
}
