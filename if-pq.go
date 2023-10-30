/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "golang.org/x/exp/constraints"

// AggregatingPriorityQueue uses cached priority obtained from
// Aggregators that operates on the values outside of the AggregatingPriorityQueue.
//   - the Update method reprioritizes an updated value element
type AggregatingPriorityQueue[V any, P constraints.Ordered] interface {
	// Get retrieves a possible value container associated with valuep
	Get(valuep *V) (aggregator Aggregator[V, P], ok bool)
	// Put stores a new value container associated with valuep
	//   - the valuep is asusmed to not have a node in the queue
	Put(valuep *V, aggregator Aggregator[V, P])
	// Update re-prioritizes a value
	Update(valuep *V)
	// Clear empties the priority queue. The hashmap is left intact.
	Clear()
	// List returns the first n or default all values by pirority
	List(n ...int) (aggregatorQueue []AggregatePriority[V, P])
}

// PriorityQueue is a pointer-identity-to-value map of updatable values traversable by rank.
//   - PriorityQueue operates directly on value by caching priority from the pritority function.
//   - the AddOrUpdate method reprioritizes an updated value element
//   - V is a value reference composite type that is comparable, ie. not slice map function.
//     Preferrably, V is interface or pointer to struct type.
//   - P is an ordered type such as Integer Floating-Point or string, used to rank the V values
//   - values are added or updated using AddOrUpdate method distinguished by
//     (computer science) identity
//   - if the same comparable value V is added again, that value is re-ranked
//   - priority P is computed from a value V using the priorityFunc function.
//     The piority function may be examining field values of a struct
//   - values can have the same rank. If they do, equal rank is provided in insertion order
//   - pqs.NewPriorityQueue[V any, P constraints.Ordered]
//   - pqs.NewRankingThreadSafe[V comparable, R constraints.Ordered](
//     ranker func(value *V) (rank R)))
type PriorityQueue[V any, P constraints.Ordered] interface {
	// AddOrUpdate adds a new value to the prioirty queue or updates the priority of a value
	// that has changed.
	AddOrUpdate(value *V)
	// List returns the first n or default all values by priority
	List(n ...int) (valueQueue []*V)
}

// AggregatePriority caches the priority value from an aggregator for priority.
//   - V is the value type used as a pointer
//   - P is the priority type descending order, ie. Integer Floating-Point string
type AggregatePriority[V any, P constraints.Ordered] interface {
	// Aggregator returns the aggregator associated with this AggregatePriority
	Aggregator() (aggregator Aggregator[V, P])
	// Update caches the current priority from the aggregator
	Update()
	// Priority returns the effective cached priority
	//	- Priority is used by consumers of the AggregatingPriorityQueue
	CachedPriority() (priority P)
	// Index indicates insertion order
	//	- used for ordering elements of equal priority
	Index() (index int)
}

// Aggregator aggregates, snapshots and assigns priority to an associated value.
//   - V is the value type used as a pointer
//   - V may be a thread-safe object whose values change in real-time
//   - P is the priority type descending order, ie. Integer Floating-Point string
type Aggregator[V any, P constraints.Ordered] interface {
	// Value returns the value object this Aggregator is associated with
	//	- the Value method is used by consumers of the AggregatingPriorityQueue
	Value() (valuep *V)
	// Aggregate aggregates and snapshots data values from the value object.
	//	- Aggregate is invoked outside of AggregatingPriorityQueue
	Aggregate()
	// Priority returns the current priority for the associated value
	//	- this priority is cached by AggregatePriority
	Priority() (priority P)
}

// AssignedPriority contains the assigned priority for a priority-queue element
//   - V is the element value type whose pointer-value provides identity
//   - P is the priority, a descending-ordered type: Integer Floating-Point string
type AssignedPriority[V any, P constraints.Ordered] interface {
	SetPriority(priority P)
}
