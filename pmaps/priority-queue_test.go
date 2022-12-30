/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Ranking is a map of updatable values traversable by rank
package pmaps

import (
	"strconv"
	"testing"

	"github.com/haraldrudell/parl"
)

func TestNewRanking(t *testing.T) {
	type entity struct{ value int }
	value1 := 1
	value2 := 2
	value3 := 3
	value4 := 4
	entity1 := entity{value: value1}
	entity2 := entity{value: value3}
	entity3 := entity{value: value2}
	entityStringMap := map[*entity]string{
		&entity1: "entity1",
		&entity2: "entity2",
		&entity3: "entity3",
	}
	ranker := func(entityN *entity) (rank int) {
		return entityN.value
	}
	exp1 := []*entity{&entity2, &entity3, &entity1}
	exp2 := []*entity{&entity1, &entity2, &entity3}
	exp3 := []*entity{&entity1}
	expLength := 3

	var ranking parl.PriorityQueue[entity, int]
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
			// strconv.Quote prints empty strings real well
			result[i] = strconv.Quote(entityStringMap[entityp])
		}
		return
	}

	ranking = NewPriorityQueue(ranker)
	if pmapsRanking, ok = ranking.(*PriorityQueue[entity, int]); !ok {
		t.Error("type asserton failed")
		t.FailNow()
	}

	// AddOrUpdate
	ranking.AddOrUpdate(&entity1)
	ranking.AddOrUpdate(&entity2)
	ranking.AddOrUpdate(&entity3)
	entity1.value = value4
	if length = len(pmapsRanking.m); length != expLength {
		t.Errorf("bad map length1 %d exp %d", length, expLength)
	}
	if pmapsRanking.queue.Length(); length != expLength {
		t.Errorf("bad Length1 %d exp %d", length, expLength)
	}
	if rankList = ranking.List(); !isSameRanking(rankList, exp1) {
		t.Errorf("bad list1 %v exp %v", makeListPrintable(rankList), makeListPrintable(exp1))
	}

	// List
	ranking.AddOrUpdate(&entity1)
	if length = len(pmapsRanking.m); length != expLength {
		t.Errorf("bad length2 %d exp %d", length, expLength)
	}
	if rankList = ranking.List(); !isSameRanking(rankList, exp2) {
		t.Errorf("bad list2 %v exp %v", makeListPrintable(rankList), makeListPrintable(exp2))
	}

	if rankList = ranking.List(1); !isSameRanking(rankList, exp3) {
		t.Errorf("bad list3 %v exp %v", makeListPrintable(rankList), makeListPrintable(exp3))
	}

	// rankNode
	entity2.value = value1
	ranking.AddOrUpdate(&entity2)
	entity1.value = value1
	ranking.AddOrUpdate(&entity1)
	entity3.value = value1
	ranking.AddOrUpdate(&entity3)
}
