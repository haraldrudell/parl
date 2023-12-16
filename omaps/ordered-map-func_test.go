/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omaps

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestNewOrderedMapFunc(t *testing.T) {
	var debug = false
	var v1 = V{1}
	var v2 = V{2}
	var v3 = V{3}
	var expList = []*V{&v1, &v3}

	// Clone() Put()
	// btreeMap: Get() Length() Range() Delete() Clear() List()
	var m OrderedMapFunc[int, *V]

	var vPointers []*V

	m = *NewOrderedMapFunc[int, *V](ascendingVvalueLess)

	// put in order 2, 3, 1
	m.Put(v2.value, &v2)
	t.Logf("%v", m.List())
	m.Put(v3.value, &v3)
	t.Logf("%v", m.List())
	m.Put(v1.value, &v1)
	t.Logf("%v", m.List())
	// delete 2
	m.Delete(v2.value)

	// List should return 1, 3
	vPointers = m.List()

	// vPointers: [1 3]
	if debug {
		var mapContents = make([]string, m.m2.Length())
		var i = 0
		m.m2.Range(func(key int, value *V) (keepGoing bool) {
			mapContents[i] = fmt.Sprintf("%d: %d", key, value.value)
			i++
			return true
		})
		slices.Sort(mapContents)
		t.Logf("map: %v", strings.Join(mapContents, "\x20"))
		t.Logf("vPointers: %v", vPointers)
		t.Fail()
	}

	if !slices.Equal(vPointers, expList) {
		t.Logf("List: %v exp %v", vPointers, expList)
	}
}

// V is a test struct with ebcapsulated integer value
type V struct{ value int }

func (v *V) String() (s string) { return strconv.Itoa(v.value) }

// sorts by the encapsulated integer value ascending
func ascendingVvalueLess(a, b *V) (aBeforeB bool) { return a.value < b.value }
