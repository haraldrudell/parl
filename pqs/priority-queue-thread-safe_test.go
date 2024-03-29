/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pqs

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

	var ranking parl.PriorityQueue[entity, int]
	var pmapsRankingThreadSafe *PriorityQueueThreadSafe[entity, int]
	var pmapsRanking *PriorityQueue[entity, int]
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

	ranking = NewPriorityQueueThreadSafe(ranker)
	ranking.AddOrUpdate(&entity1)
	ranking.AddOrUpdate(&entity2)
	pmapsRankingThreadSafe, ok = ranking.(*PriorityQueueThreadSafe[entity, int])
	if !ok {
		t.Errorf("type assertion failed: ranking.(*RankingThreadSafe[entity, int])")
		t.FailNow()
	}
	pmapsRanking, ok = pmapsRankingThreadSafe.PriorityQueue.(*PriorityQueue[entity, int])
	if !ok {
		t.Errorf("type assertion failed: pmapsRankingThreadSafe.Ranking.(*Ranking[entity, int])")
		t.FailNow()
	}
	length = pmapsRanking.queue.Length()
	if length != expLength {
		t.Errorf("bad Length1 %d exp %d", length, expLength)
	}
	if rankList = ranking.List(); !isSameRanking(rankList, exp1) {
		t.Errorf("bad list1 %v exp %v", makeListPrintable(rankList), makeListPrintable(exp1))
	}
}
