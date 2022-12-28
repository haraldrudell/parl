/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// RankingThreadSafe is a thread-safe pointer-identity-to-value map of updatable values traversable by rank.
// RankingThreadSafe implements [parl.Ranking][V comparable, R constraints.Ordered].
package pmaps

import (
	"sync"

	"github.com/haraldrudell/parl"
	"golang.org/x/exp/constraints"
)

// RankingThreadSafe is a thread-safe pointer-identity-to-value map of updatable values traversable by rank.
// RankingThreadSafe implements [parl.Ranking][V comparable, R constraints.Ordered].
//   - V is a value reference composite type that is comparable, ie. not slice map function.
//     Preferrably, V is interface or pointer to struct type.
//   - R is an ordered type such as int floating-point string, used to rank the V values
//   - values are added or updated using AddOrUpdate method distinguished by
//     (computer science) identity
//   - if the same comparable value V is added again, that value is re-ranked
//   - rank R is computed from a value V using the ranker function.
//     The ranker function may be examining field values of a struct
//   - values can have the same rank. If they do, equal rank is provided in insertion order
type RankingThreadSafe[V comparable, R constraints.Ordered] struct {
	lock sync.RWMutex
	parl.Ranking[V, R]
}

// NewRanking returns a thread-safe map of updatable values traversable by rank
func NewRankingThreadSafe[V comparable, R constraints.Ordered](
	ranker func(value *V) (rank R),
) (o1 parl.Ranking[V, R]) {
	return &RankingThreadSafe[V, R]{
		Ranking: NewRanking[V, R](ranker),
	}
}

// AddOrUpdate adds a new value to the ranking or updates the ranking of a value
// that has changed.
func (mp *RankingThreadSafe[V, R]) AddOrUpdate(value *V) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.Ranking.AddOrUpdate(value)
}

// List returns the first n or default all values by rank
func (mp *RankingThreadSafe[V, R]) List(n ...int) (list []*V) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.Ranking.List(n...)
}
