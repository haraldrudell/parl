/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"strconv"
	"testing"

	"github.com/haraldrudell/parl"
)

func TestNewOrderedThreadSafe(t *testing.T) {
	type entity struct{ value int }
	value1 := 1
	value2 := 2
	entity1 := entity{value: value1}
	entity2 := entity{value: value2}
	entityStringMap := map[*entity]string{
		&entity1: "entity1",
		&entity2: "entity2",
	}
	ranker := func(entityN *entity) (rank int) {
		return entityN.value
	}
	exp1 := []*entity{&entity2, &entity1}
	expLength := 2

	var ranking parl.Ranking[entity, int]
	var pmapsRankingThreadSafe *RankingThreadSafe[entity, int]
	var pmapsRanking *Ranking[entity, int]
	var ok bool
	var rankList []*entity
	var length int
	isSameRanking := func(a, b []*entity) (isSame bool) {
		if len(a) != len(b) {
			return
		}
		for i, entityp := range a {
			if entityp != b[i] {
				return
			}
		}
		return true
	}
	makeListPrintable := func(entities []*entity) (result []string) {
		result = make([]string, len(entities))
		for i, entityp := range entities {
			result[i] = strconv.Quote(entityStringMap[entityp])
		}
		return
	}

	ranking = NewRankingThreadSafe(ranker)
	ranking.AddOrUpdate(&entity1)
	ranking.AddOrUpdate(&entity2)
	pmapsRankingThreadSafe, ok = ranking.(*RankingThreadSafe[entity, int])
	if !ok {
		t.Errorf("type assertion failed: ranking.(*RankingThreadSafe[entity, int])")
		t.FailNow()
	}
	pmapsRanking, ok = pmapsRankingThreadSafe.Ranking.(*Ranking[entity, int])
	if !ok {
		t.Errorf("type assertion failed: pmapsRankingThreadSafe.Ranking.(*Ranking[entity, int])")
		t.FailNow()
	}
	length = pmapsRanking.ranking.Length()
	if length != expLength {
		t.Errorf("bad Length1 %d exp %d", length, expLength)
	}
	if rankList = ranking.List(); !isSameRanking(rankList, exp1) {
		t.Errorf("bad list1 %v exp %v", makeListPrintable(rankList), makeListPrintable(exp1))
	}
}
