/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"github.com/haraldrudell/parl/parli"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/constraints"
)

type AggregatePriority[V any, P constraints.Ordered] struct {
	assignedPriority AssignedPriority[V, P]
	aggregator       parli.Aggregator[V, P]
}

func NewAggregatePriority[V any, P constraints.Ordered](
	value *V,
	index int,
	aggregator parli.Aggregator[V, P],
) (aggregatePriority parli.AggregatePriority[V, P]) {
	return &AggregatePriority[V, P]{
		assignedPriority: *NewAssignedPriority(aggregator.Priority(), index, value),
		aggregator:       aggregator,
	}
}

func (ap *AggregatePriority[V, P]) Aggregator() (aggregator parli.Aggregator[V, P]) {
	return ap.aggregator
}

func (ap *AggregatePriority[V, P]) Update() {
	ap.assignedPriority.SetPriority(ap.aggregator.Priority())
}

func (ap *AggregatePriority[V, P]) CachedPriority() (priority P) {
	return ap.assignedPriority.Priority
}

func (a *AggregatePriority[V, P]) Cmp(b parli.AggregatePriority[V, P]) (result int) {
	var b1 *AggregatePriority[V, P]
	var ok bool
	if b1, ok = b.(*AggregatePriority[V, P]); !ok {
		panic(perrors.ErrorfPF("Comparison with bad type: %T expected: %T", b, a))
	}
	return a.assignedPriority.Cmp(&b1.assignedPriority)
}
