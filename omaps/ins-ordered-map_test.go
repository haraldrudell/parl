/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omaps

import (
	"slices"
	"testing"
)

func TestInsOrderedMap(t *testing.T) {
	var keys = []int{10, 9, 8}
	var values = []int{3, 2, 1}
	var deleteKeyIndex = 1
	var updateKeyIndex = 2
	var updatedValue = 4
	//	- first keys and values are inserted: 3 2 1
	//	- then value: 2 is deleted
	//	- then value: 1 is updated to 4, not affecting order
	//	- result is 3, 4
	var expList = []int{values[0], updatedValue}

	// populate map
	var m = NewInsOrderedMap[int, int]()
	for i, k := range keys {
		m.Put(k, values[i])
	}
	m.Delete(keys[deleteKeyIndex])
	m.Put(keys[updateKeyIndex], updatedValue)

	// get list
	var list = m.List()

	if !slices.Equal(list, expList) {
		t.Logf("map: %v", m.m)
		t.Logf("tree: %d", m.tree.Len())
		t.Errorf("bad list:\n%v exp\n%v",
			list, expList,
		)
	}
}
