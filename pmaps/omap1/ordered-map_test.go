/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omap1

import (
	"maps"
	"slices"
	"strconv"
	"testing"
)

func TestOrderedMapMove(t *testing.T) {
	// MoveAfter()
	// MoveBefore()
	// KeyStrings()
	const (
		key1, key2, key3, key4 = 1, 2, 3, 4
		value1, value2, value3 = 11, 12, 13
	)
	var (
		keys     = []int{key1, key2, key3}
		values   = []int{value1, value2, value3}
		mappings = func() (m []Mapping[int, int]) {
			m = make([]Mapping[int, int], len(keys))
			for i := range keys {
				m[i] = Mapping[int, int]{Key: keys[i], Value: values[i]}
			}
			return
		}()
		// 2, 1, 3
		expKeys       = []int{key2, key1, key3}
		expKeysBefore = []int{key1, key3, key2}
		// 2, 1, 3
		expKeysStrings = []string{strconv.Itoa(key2), strconv.Itoa(key1), strconv.Itoa(key3)}
	)

	var (
		m     = MakeOrderedMapFromMappings(mappings)
		reset = func() {
			m = MakeOrderedMapFromMappings(mappings)
		}
		didMove   bool
		keyStings []string
	)

	// MoveAfter()
	reset()
	didMove = m.MoveAfter(key1, key1)
	if didMove {
		t.Error("FAIL didMove true")
	}
	didMove = m.MoveAfter(key4, key1)
	if didMove {
		t.Error("FAIL didMove true")
	}
	didMove = m.MoveAfter(key1, key4)
	if didMove {
		t.Error("FAIL didMove true")
	}
	didMove = m.MoveAfter(key1, key2)
	if !didMove {
		t.Error("FAIL didMove false")
	}
	keys = m.Keys()
	if !slices.Equal(keys, expKeys) {
		t.Errorf("FAIL keys %v exp %v", keys, expKeys)
	}
	keyStings = m.KeyStrings()
	if !slices.Equal(keyStings, expKeysStrings) {
		t.Errorf("FAIL keys %v exp %v", keys, expKeysStrings)
	}

	// MoveBefore()
	reset()
	didMove = m.MoveBefore(key1, key1)
	if didMove {
		t.Error("FAIL didMove true")
	}
	didMove = m.MoveBefore(key4, key1)
	if didMove {
		t.Error("FAIL didMove true")
	}
	didMove = m.MoveBefore(key1, key4)
	if didMove {
		t.Error("FAIL didMove true")
	}
	didMove = m.MoveBefore(key3, key2)
	if !didMove {
		t.Error("FAIL didMove false")
	}
	keys = m.Keys()
	if !slices.Equal(keys, expKeysBefore) {
		t.Errorf("FAIL keys %v exp %v", keys, expKeys)
	}

}

func TestOrderedMap(t *testing.T) {
	// GetAndMoveToBack()
	// GetAndMoveToFront()
	// Newest()
	// Oldest()
	const (
		key1, key2, key3, key4 = 1, 2, 3, 4
		value1, value2, value3 = 11, 12, 13
	)
	var (
		keys     = []int{key1, key2, key3}
		values   = []int{value1, value2, value3}
		mappings = func() (m []Mapping[int, int]) {
			m = make([]Mapping[int, int], len(keys))
			for i := range keys {
				m[i] = Mapping[int, int]{Key: keys[i], Value: values[i]}
			}
			return
		}()
		expKeys      = []int{key1, key3, key2}
		expKeysFront = []int{key2, key1, key3}
	)

	var (
		m      = MakeOrderedMapFromMappings(mappings)
		mEmpty = MakeOrderedMap[int, int]()
		reset  = func() {
			m = MakeOrderedMapFromMappings(mappings)
		}
		key      int
		value    int
		hasValue bool
	)

	// Newest()
	key, value, hasValue = m.Newest()
	if key != key3 {
		t.Errorf("FAIL key %d exp %d", key, key3)
	}
	if value != value3 {
		t.Errorf("FAIL value %d exp %d", value, value3)
	}
	if !hasValue {
		t.Error("FAIL hasValue false")
	}
	key, value, hasValue = mEmpty.Newest()
	if key != 0 {
		t.Errorf("FAIL key %d exp %d", key, 0)
	}
	if value != 0 {
		t.Errorf("FAIL value %d exp %d", value, 0)
	}
	if hasValue {
		t.Error("FAIL hasValue true")
	}

	// Oldest()
	key, value, hasValue = m.Oldest()
	if key != key1 {
		t.Errorf("FAIL key %d exp %d", key, key1)
	}
	if value != value1 {
		t.Errorf("FAIL value %d exp %d", value, value1)
	}
	if !hasValue {
		t.Error("FAIL hasValue false")
	}
	key, value, hasValue = mEmpty.Oldest()
	if key != 0 {
		t.Errorf("FAIL key %d exp %d", key, 0)
	}
	if value != 0 {
		t.Errorf("FAIL value %d exp %d", value, 0)
	}
	if hasValue {
		t.Error("FAIL hasValue true")
	}

	// GetAndMoveToBack()
	reset()
	value, hasValue = m.GetAndMakeNewest(key2)
	if value != value2 {
		t.Errorf("FAIL value %d exp %d", value, value2)
	}
	if !hasValue {
		t.Error("FAIL hasValue false")
	}
	keys = m.Keys()
	if !slices.Equal(keys, expKeys) {
		t.Errorf("FAIL keys %v exp %v", keys, expKeys)
	}
	value, hasValue = mEmpty.GetAndMakeNewest(key2)
	if value != 0 {
		t.Errorf("FAIL value %d exp %d", value, 0)
	}
	if hasValue {
		t.Error("FAIL hasValue true")
	}
	value, hasValue = m.GetAndMakeNewest(key4)
	if value != 0 {
		t.Errorf("FAIL value %d exp %d", value, 0)
	}
	if hasValue {
		t.Error("FAIL hasValue true")
	}

	// GetAndMoveToFront()
	reset()
	value, hasValue = m.GetAndMakeOldest(key2)
	if value != value2 {
		t.Errorf("FAIL value %d exp %d", value, value2)
	}
	if !hasValue {
		t.Error("FAIL hasValue false")
	}
	keys = m.Keys()
	if !slices.Equal(keys, expKeysFront) {
		t.Errorf("FAIL keys %v exp %v", keys, expKeysFront)
	}
	value, hasValue = mEmpty.GetAndMakeOldest(key2)
	if value != 0 {
		t.Errorf("FAIL value %d exp %d", value, 0)
	}
	if hasValue {
		t.Error("FAIL hasValue true")
	}
	value, hasValue = m.GetAndMakeOldest(key4)
	if value != 0 {
		t.Errorf("FAIL value %d exp %d", value, 0)
	}
	if hasValue {
		t.Error("FAIL hasValue true")
	}

}

func TestGetPut(t *testing.T) {
	// Clear()
	// Clone()
	// Delete()
	// Get()
	// Put()
	// Contains()
	const (
		key1, key2, key3       = 1, 2, 3
		value1, value2, value3 = 11, 12, 13
		expLength              = 1
	)
	var (
		m          = MakeOrderedMap[int, int]()
		old        int
		hadMapping bool
		value      int
		hasValue   bool
		m2         OrderedMap[int, int]
		didClear   bool
		contains   bool
	)

	// Put()
	old, hadMapping = m.Put(key1, value1)
	if old != 0 {
		t.Errorf("FAIL old %d exp %d", old, 0)
	}
	if hadMapping {
		t.Error("FAIL hadMapping true")
	}
	old, hadMapping = m.Put(key2, value2)
	if old != 0 {
		t.Errorf("FAIL old %d exp %d", old, 0)
	}
	if hadMapping {
		t.Error("FAIL hadMapping true")
	}
	old, hadMapping = m.Put(key2, value3)
	if old != value2 {
		t.Errorf("FAIL old %d exp %d", old, value2)
	}
	if !hadMapping {
		t.Error("FAIL hadMapping false")
	}

	// Get()
	value, hasValue = m.Get(key2)
	if value != value3 {
		t.Errorf("FAIL value %d exp %d", value, value3)
	}
	if !hasValue {
		t.Error("FAIL hasValue false")
	}
	value, hasValue = m.Get(key3)
	if value != 0 {
		t.Errorf("FAIL value %d exp %d", value, 0)
	}
	if hasValue {
		t.Error("FAIL hasValue true")
	}

	// Contains()
	contains = m.Contains(key2)
	if !contains {
		t.Error("FAIL contains false")
	}
	contains = m.Contains(key3)
	if contains {
		t.Error("FAIL contains true")
	}

	// Delete()
	old, hadMapping = m.Delete(key3)
	if old != 0 {
		t.Errorf("FAIL old %d exp %d", old, 0)
	}
	if hadMapping {
		t.Error("FAIL hadMapping true")
	}
	old, hadMapping = m.Delete(key1)
	if old != value1 {
		t.Errorf("FAIL old %d exp %d", old, value1)
	}
	if !hadMapping {
		t.Error("FAIL hadMapping false")
	}

	// Clone()
	m2 = m.Clone()
	if m2.Length() != expLength {
		t.Errorf("FAIL Length %d exp %d", m2.Length(), expLength)
	}

	// Clear
	didClear = m.Clear()
	if !didClear {
		t.Error("didClear false")
	}
	didClear = m.Clear()
	if didClear {
		t.Error("didClear true")
	}

	if m2.Length() == 0 {
		t.Errorf("FAIL m2 length zero")
	}
}

func TestTraverse(t *testing.T) {
	// Traverse() TraverseBackwards()
	var (
		keys        = []int{1, 2, 3}
		values      = []int{11, 12, 13}
		middleIndex = 1
		expMap      = func() (m map[int]int) {
			m = make(map[int]int, len(keys))
			for i := range len(keys) {
				m[keys[i]] = values[i]
			}
			return
		}()
		mappings = func() (m []Mapping[int, int]) {
			m = make([]Mapping[int, int], len(keys))
			for i := range keys {
				m[i] = Mapping[int, int]{Key: keys[i], Value: values[i]}
			}
			return
		}()
		expMapForward = map[int]int{
			keys[1]: values[1],
			keys[2]: values[2],
		}
		expMapBackwards = map[int]int{
			keys[1]: values[1],
			keys[0]: values[0],
		}
	)
	var (
		m     = MakeOrderedMapFromMappings(mappings)
		goMap = make(map[int]int)
		i     int
	)

	clear(goMap)
	i = 0
	for key, value := range m.Traverse() {
		if key != keys[i] {
			t.Errorf("FAIL key%d %d exp %d", i, key, keys[i])
		}
		i++
		goMap[key] = value
	}
	if !maps.Equal(goMap, expMap) {
		t.Errorf("FAIL map %v exp %v", goMap, expMap)
	}

	clear(goMap)
	i = middleIndex
	for key, value := range m.Traverse(keys[middleIndex]) {
		if key != keys[i] {
			t.Errorf("FAIL key%d %d exp %d", i, key, keys[i])
		}
		i++
		goMap[key] = value
	}
	if !maps.Equal(goMap, expMapForward) {
		t.Errorf("FAIL map %v exp %v", goMap, expMapForward)
	}

	clear(goMap)
	i = len(keys) - 1
	for key, value := range m.TraverseBackwards() {
		if key != keys[i] {
			t.Errorf("FAIL key%d %d exp %d", i, key, keys[i])
		}
		i--
		goMap[key] = value
	}
	if !maps.Equal(goMap, expMap) {
		t.Errorf("FAIL map %v exp %v", goMap, expMap)
	}

	clear(goMap)
	i = middleIndex
	for key, value := range m.TraverseBackwards(keys[middleIndex]) {
		if key != keys[i] {
			t.Errorf("FAIL key%d %d exp %d", i, key, keys[i])
		}
		i--
		goMap[key] = value
	}
	if !maps.Equal(goMap, expMapBackwards) {
		t.Errorf("FAIL map %v exp %v", goMap, expMapBackwards)
	}

}

func TestMakeOrderedMap(t *testing.T) {
	// MakeOrderedMap()
	// MakeOrderedMapFromKeys()
	// MakeOrderedMapFromMappings()
	// also tests Keys() GoMap() Length()
	const (
		key1, key2     = 1, 2
		value1, value2 = 11, 12
		expLength      = 2
	)
	var (
		keys     = []int{key1, key2}
		mappings = []Mapping[int, int]{{
			Key: key1, Value: value1,
		}, {
			Key: key2, Value: value2,
		}}
		expMap = map[int]int{
			key1: value1,
			key2: value2,
		}
	)
	var (
		m          OrderedMap[int, int]
		actualKeys []int
		goMap      map[int]int
	)
	m = MakeOrderedMap[int, int]()
	_ = m
	m = MakeOrderedMap[int, int](1)
	_ = m
	m = MakeOrderedMapFromKeys[int, int](keys)
	if m.Length() != expLength {
		t.Errorf("Fail bad length %d exp %d", m.Length(), expLength)
	}
	actualKeys = m.Keys()
	if !slices.Equal(actualKeys, keys) {
		t.Errorf("Fail bad keys %v exp %v", actualKeys, keys)
	}
	m = MakeOrderedMapFromMappings(mappings)
	actualKeys = m.Keys()
	if !slices.Equal(actualKeys, keys) {
		t.Errorf("Fail bad keys %v exp %v", actualKeys, keys)
	}
	goMap = m.GoMap()
	if !maps.Equal(goMap, expMap) {
		t.Errorf("Fail bad maps %v exp %v", goMap, expMap)
	}
}
