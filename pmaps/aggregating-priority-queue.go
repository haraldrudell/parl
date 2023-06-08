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

// AggregatingPriorityQueue implements a priority queue using cached priorities from aggregators
//   - identity is the pointer value to each aggregator of type V
//   - P is the type used for priority, ordered highest first
//   - insertion order is used for equal priorities, order lowest/earliest first
type AggregatingPriorityQueue[V any, P constraints.Ordered] struct {
	// queue is a list of queue nodes ordered by rank
	//	- element is parli.AggregatePriority: pointer to AggregatePriority
	//	- AggregatePriority holds a cached priority value and the aggregator with running totals
	queue parli.Ordered[parli.AggregatePriority[V, P]]
	// m provides O(1) access to priority data-nodes via the value-pointer
	m map[*V]parli.AggregatePriority[V, P]
	// indexGenerator provides IDs for insertion-ordering
	indexGenerator ids.UniqueIDint
	// non-zero for limiting queue length
	maxQueueLength int
}

var _ AggregatePriority[int, int]

// NewAggregatingPriorityQueue returns a map of updatable values traversable by rank
func NewAggregatingPriorityQueue[V any, P constraints.Ordered](
	maxQueueLength ...int,
) (priorityQueue parli.AggregatingPriorityQueue[V, P]) {
	var maxQueueLength0 int
	if len(maxQueueLength) > 0 {
		maxQueueLength0 = maxQueueLength[0]
	}
	if maxQueueLength0 < 1 {
		maxQueueLength0 = 1
	}
	var a *AggregatePriority[V, P]
	return &AggregatingPriorityQueue[V, P]{
		m:              map[*V]parli.AggregatePriority[V, P]{},
		queue:          pslices.NewOrderedAny(a.Cmp),
		maxQueueLength: maxQueueLength0,
	}
}

// Get retrieves a the value container with running totals associated with the identity valuep
func (a *AggregatingPriorityQueue[V, P]) Get(valuep *V) (aggregator parli.Aggregator[V, P], ok bool) {
	var nodep parli.AggregatePriority[V, P]
	if nodep, ok = a.m[valuep]; ok {
		aggregator = nodep.Aggregator()
	}
	return
}

// Put stores a new value container associated with valuep
//   - the valuep is assumed to not have a node in the queue
func (a *AggregatingPriorityQueue[V, P]) Put(valuep *V, aggregator parli.Aggregator[V, P]) {

	// create aggregatePriority with current priority from aggregator
	aggregatePriority := NewAggregatePriority(
		valuep,
		a.indexGenerator.ID(),
		aggregator,
	)

	// store in map
	a.m[valuep] = aggregatePriority
	a.insert(aggregatePriority)
}

// Update re-prioritizes a value
func (a *AggregatingPriorityQueue[V, P]) Update(valuep *V) {

	var aggregatePriority parli.AggregatePriority[V, P]
	var ok bool
	if aggregatePriority, ok = a.m[valuep]; !ok {
		return // value priority does not exist return
	}

	// update order: that’s what we do here. We keep order
	a.queue.Delete(aggregatePriority)
	var _ = (&AggregatePriority[int, int]{}).Update
	aggregatePriority.Update()
	a.insert(aggregatePriority)
}

var _ = ((parli.AggregatingPriorityQueue[int, int])(&AggregatingPriorityQueue[int, int]{})).Clear

// Clear empties the priority queue. The hashmap is left intact.
func (a *AggregatingPriorityQueue[V, P]) Clear() {
	a.queue.Clear()
}

// List returns the first n or default all values by pirority
func (a *AggregatingPriorityQueue[V, P]) List(n ...int) (aggregatorQueue []parli.AggregatePriority[V, P]) {
	return a.queue.List(n...)
}

func (a *AggregatingPriorityQueue[V, P]) insert(aggregatePriority parli.AggregatePriority[V, P]) {

	// enforce max length
	if a.maxQueueLength > 0 && a.queue.Length() == a.maxQueueLength {
		var ap *AggregatePriority[V, P]
		if ap.Cmp(aggregatePriority, a.queue.Element(a.maxQueueLength-1)) >= 0 {
			return // aggregatePriority has too low priority to make the cut
		}
		a.queue.DeleteIndex(a.maxQueueLength - 1)
	}

	// update order: that’s what we do here. We keep order
	a.queue.Insert(aggregatePriority)
}
