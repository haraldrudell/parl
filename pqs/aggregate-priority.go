/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// pqs provides legacy priority-queue implementation likely to be deprecated
package pqs

import (
	"github.com/haraldrudell/parl"
	"golang.org/x/exp/constraints"
)

// AggregatePriority provides cached priority values and order function
type AggregatePriority[V any, P constraints.Ordered] struct {
	assignedPriority AssignedPriority[V, P]
	aggregator       parl.Aggregator[V, P]
}

// NewAggregatePriority returns an object providing cached priority values and order function
func NewAggregatePriority[V any, P constraints.Ordered](
	value *V,
	index int,
	aggregator parl.Aggregator[V, P],
) (aggregatePriority parl.AggregatePriority[V, P]) {
	return &AggregatePriority[V, P]{
		assignedPriority: *NewAssignedPriority(aggregator.Priority(), index, value),
		aggregator:       aggregator,
	}
}

var _ = ((parl.AggregatePriority[int, int])(&AggregatePriority[int, int]{})).Aggregator

// Aggregator returns the object calculating values
func (a *AggregatePriority[V, P]) Aggregator() (aggregator parl.Aggregator[V, P]) {
	return a.aggregator
}

var _ = ((parl.AggregatePriority[int, int])(&AggregatePriority[int, int]{})).Update

// Update caches the current priority from the aggregator
func (a *AggregatePriority[V, P]) Update() {
	a.assignedPriority.SetPriority(a.aggregator.Priority())
}

var _ = ((parl.AggregatePriority[int, int])(&AggregatePriority[int, int]{})).CachedPriority

// Priority returns the effective cached priority
func (a *AggregatePriority[V, P]) CachedPriority() (priority P) {
	return a.assignedPriority.Priority
}

// Priority returns the effective cached priority
func (a *AggregatePriority[V, P]) Index() (index int) {
	return a.assignedPriority.Index
}

// Cmp returns a comparison of two AggregatePriority objects that represents value elements.
//   - Cmp is a custom comparison function to be used with pslices and slices packages
//   - Cmp makes AggregatePriority ordered
//   - the Priority used is uncached value
func (x *AggregatePriority[V, P]) Cmp(a, b parl.AggregatePriority[V, P]) (result int) {
	aPriority := a.CachedPriority()
	bPriority := b.CachedPriority()
	if aPriority > bPriority { // highest priority first
		return -1
	} else if aPriority < bPriority {
		return 1
	}
	aIndex := a.Index()
	bIndex := b.Index()
	if aIndex < bIndex { // lowest index first
		return -1
	} else if aIndex > bIndex {
		return 1
	}
	return 0
}
