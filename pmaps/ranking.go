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

// Ranking is a pointer-identity-to-value map of updatable values traversable by rank.
// Ranking implements [parl.Ranking][V comparable, R constraints.Ordered].
//   - V is a value reference composite type that is comparable, ie. not slice map function.
//     Preferrably, V is interface or pointer to struct type.
//   - R is an ordered type such as int floating-point string, used to rank the V values
//   - values are added or updated using AddOrUpdate method distinguished by
//     (computer science) identity
//   - if the same comparable value V is added again, that value is re-ranked
//   - rank R is computed from a value V using the ranker function.
//     The ranker function may be examining field values of a struct
//   - values can have the same rank. If they do, equal rank is provided in insertion order
type Ranking[V any, R constraints.Ordered] struct {
	// ranker is the function computing rank for a value-pointer
	ranker func(value *V) (key R)
	// ranking is a list of ranking nodes ordered by rank
	ranking parl.Ordered[*rankingNode[V, R]]
	// m is a map providing O(1) access to ranking nodes by value-pointer
	m map[*V]*rankingNode[V, R]
}

// rankingNode is an internal value structure used by Ranking
type rankingNode[V any, R constraints.Ordered] struct {
	Rank  R
	Index int
	Value *V
}

// NewRanking returns a map of updatable values traversable by rank
func NewRanking[V comparable, R constraints.Ordered](
	ranker func(value *V) (rank R),
) (ranking parl.Ranking[V, R]) {
	if ranker == nil {
		perrors.NewPF("ranker cannot be nil")
	}
	r := Ranking[V, R]{
		ranker: ranker,
		m:      map[*V]*rankingNode[V, R]{},
	}
	r.ranking = pslices.NewOrderedAny[*rankingNode[V, R]](r.rankNode)
	return &r
}

// AddOrUpdate adds a new value to the ranking or updates the ranking of a value
// that has changed.
func (ra *Ranking[V, R]) AddOrUpdate(valuep *V) {

	// get rank and identity
	rank := ra.ranker(valuep)

	var np *rankingNode[V, R]
	var ok bool
	if np, ok = ra.m[valuep]; !ok {

		// new value case
		np = &rankingNode[V, R]{Rank: rank, Index: ra.ranking.Length(), Value: valuep}
		ra.m[valuep] = np
	} else {

		// node update case
		ra.ranking.Delete(np)
		np.Rank = rank
	}

	// update order: that’s what we do here. We keep order
	ra.ranking.Insert(np)
}

// List returns the first n or default all values by rank
func (ra *Ranking[V, K]) List(n ...int) (rank []*V) {

	// get number of items n0
	var n0 int
	length := ra.ranking.Length()
	if len(n) > 0 {
		n0 = n[0]
	}
	if n0 < 1 || n0 > length {
		n0 = length
	}

	// build list
	rank = make([]*V, n0)
	rankIn := ra.ranking.List()
	for i := 0; i < n0; i++ {
		rank[i] = rankIn[i].Value
	}
	return
}

// cmp sorts descending: -1 results appears first
func (ra *Ranking[V, R]) rankNode(a, b *rankingNode[V, R]) (result int) {
	if a.Rank > b.Rank {
		return -1
	} else if a.Rank < b.Rank {
		return 1
	} else if a.Index < b.Index {
		return -1
	} else if a.Index > b.Index {
		return 1
	}
	return 0
}
