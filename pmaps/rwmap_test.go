/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// RWMap is a thread-safe mapping.
package pmaps

import (
	"context"
	"encoding/base64"
	"math/rand"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/haraldrudell/parl/parli"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/maps"
)

func TestRWMap(t *testing.T) {
	const (
		key1, key2     = "key1", "key2"
		value1, value2 = 1, 2
		// expected map length 1
		lengthExp              = 1
		lengthExpRange2        = 1
		lengthExp0             = 0
		keysLength, listLength = 1, 1
	)
	var (
		mapExp1      = map[string]int{key1: value1}
		mapExp2      = map[string]int{key1: value1, key2: value2}
		mapExpValue2 = map[string]int{key1: value2}
		mapExpKey2   = map[string]int{key2: value2}
		keysExp      = []string{key1}
		value2x      = value2
		newV         = func() (valuep *int) { return &value2x }
		makeV        = func() (value int) { return value2 }
		putIfTrue    = func(value int) (doPut bool) {
			if value != value1 {
				panic(perrors.NewPF("putif bad value"))
			}
			return true
		}
		putIfFalse = func(value int) (doPut bool) {
			if value != value1 {
				panic(perrors.NewPF("putif bad value"))
			}
			return
		}
	)

	var (
		lengthAct, value, zeroValue    int
		hasValue, rangedAll, wasNewKey bool
		ranger                         *mapRanger[string, int]
		mapAct                         map[string]int
		keys                           []string
		values                         []int
		clone                          parli.ThreadSafeMap[string, int]
		rwmap2                         *RWMap[string, int]
		goMap                          map[string]int
	)

	// Get() Put() Delete() Length() Range()
	// GetOrCreate() PutIf()
	// Clear() Clone() Clone2()
	// List() Keys()
	var rwmap *RWMap[string, int]
	var reset = func() {
		rwmap = NewRWMap2[string, int]()
		rwmap.Put(key1, value1)
	}
	var getMap = func() (mapAct map[string]int) {
		var r = newMapRanger[string, int](true)
		rwmap.Range(r.rangeFunc)
		mapAct = r.M
		return
	}

	// Length should return length
	reset()
	lengthAct = rwmap.Length()
	if lengthAct != lengthExp {
		t.Errorf("Length %d exp %d", lengthAct, lengthExp)
	}

	// Get existing key should return value
	reset()
	value, hasValue = rwmap.Get(key1)
	if !hasValue {
		t.Error("Get1 hasValue false")
	}
	if value != value1 {
		t.Errorf("Get1 value %d exp %d", value, value1)
	}

	// Get non-existing key should return no value
	reset()
	value, hasValue = rwmap.Get(key2)
	if hasValue {
		t.Error("Get2 hasValue true")
	}
	if value != zeroValue {
		t.Errorf("Get2 value %d exp %d", value, zeroValue)
	}

	// Put new key should grow map
	reset()
	rwmap.Put(key2, value2)
	mapAct = getMap()
	if !maps.Equal(mapAct, mapExp2) {
		t.Errorf("Put1 %v exp %v", mapAct, mapExp2)
	}

	// Put existing key should change map
	reset()
	rwmap.Put(key1, value2)
	mapAct = getMap()
	if !maps.Equal(mapAct, mapExpValue2) {
		t.Errorf("Put2 %v exp %v", mapAct, mapExpValue2)
	}

	// Delete existing key should change map
	reset()
	rwmap.Put(key2, value2)
	rwmap.Delete(key1)
	mapAct = getMap()
	if !maps.Equal(mapAct, mapExpKey2) {
		t.Errorf("Delete %v exp %v", mapAct, mapExpKey2)
	}

	// Range should return entire map
	reset()
	ranger = newMapRanger[string, int](true)
	rangedAll = rwmap.Range(ranger.rangeFunc)
	if !rangedAll {
		t.Error("Range1 rangedAll false")
	}
	if !maps.Equal(ranger.M, mapExp1) {
		t.Errorf("Range1 %v exp %v", ranger.M, mapExp1)
	}

	// Range can be aborted
	reset()
	rwmap.Put(key2, value2)
	ranger = newMapRanger[string, int](false)
	rangedAll = rwmap.Range(ranger.rangeFunc)
	if rangedAll {
		t.Error("Range2 rangedAll true")
	}
	if len(ranger.M) != lengthExpRange2 {
		t.Errorf("Range2 length %d exp %d", len(ranger.M), lengthExpRange2)
	}

	// Clear should empty map
	reset()
	rwmap.Clear()
	if le := rwmap.Length(); le != lengthExp0 {
		t.Errorf("Clear length %d exp %d", le, lengthExp0)
	}

	// Keys should return keys
	reset()
	keys = rwmap.Keys()
	if !slices.Equal(keys, keysExp) {
		t.Errorf("Keys1 %v exp %v", keys, keysExp)
	}

	// Keys can retrieve partial
	reset()
	rwmap.Put(key2, value2)
	keys = rwmap.Keys(keysLength)
	if len(keys) != keysLength {
		t.Fatalf("Keys2 length %d exp %d", len(keys), keysLength)
	}
	if keys[0] != key1 && keys[0] != key2 {
		t.Error("Keys2 bad keys")
	}

	// List should return values
	reset()
	values = rwmap.List()
	if !slices.Equal(values, []int{value1}) {
		t.Errorf("List1 %v exp %v", values, []int{value1})
	}

	// List can retrieve partial
	reset()
	rwmap.Put(key2, value2)
	values = rwmap.List(listLength)
	if len(values) != listLength {
		t.Fatalf("List2 length %d exp %d", len(values), listLength)
	}
	if values[0] != value1 && values[0] != value2 {
		t.Error("List2 bad keys")
	}

	// Clone should clone
	reset()
	clone = rwmap.Clone()
	if le := clone.Length(); le != lengthExp {
		t.Errorf("Clone length %d exp %d", le, lengthExp)
	}

	// Clone2 should clone
	reset()
	rwmap2 = rwmap.Clone2()
	ranger = newMapRanger[string, int](true)
	rwmap2.Range(ranger.rangeFunc)
	if !maps.Equal(ranger.M, mapExp1) {
		t.Errorf("Clone2 %v exp %v", ranger.M, mapExp1)
	}

	// Clone to Go map
	reset()
	rwmap.Clone(&goMap)
	if !maps.Equal(goMap, mapExp1) {
		t.Errorf("Clone2 %v exp %v", goMap, mapExp1)
	}

	// GetOrCreate unknown key should return nil
	reset()
	value, hasValue = rwmap.GetOrCreate(key2, nil, nil)
	if hasValue {
		t.Error("GetOrCreate1 hasValue true")
	}
	if value != zeroValue {
		t.Errorf("GetOrCreate1 value %d exp %d", value, zeroValue)
	}

	// GetOrCreate known key should return value
	reset()
	value, hasValue = rwmap.GetOrCreate(key1, nil, nil)
	if !hasValue {
		t.Error("GetOrCreate2 hasValue false")
	}
	if value != value1 {
		t.Errorf("GetOrCreate2 value %d exp %d", value, value1)
	}

	// GetOrCreate should use newV
	reset()
	value, hasValue = rwmap.GetOrCreate(key2, newV, nil)
	if !hasValue {
		t.Error("GetOrCreate3 hasValue false")
	}
	if value != value2 {
		t.Errorf("GetOrCreate3 value %d exp %d", value, value2)
	}

	// GetOrCreate should use makeV
	reset()
	value, hasValue = rwmap.GetOrCreate(key2, nil, makeV)
	if !hasValue {
		t.Error("GetOrCreate4 hasValue false")
	}
	if value != value2 {
		t.Errorf("GetOrCreate4 value %d exp %d", value, value2)
	}

	// Putif new key should add mapping
	reset()
	wasNewKey = rwmap.PutIf(key2, value2, nil)
	if !wasNewKey {
		t.Error("PutIf1 wasNewKey false")
	}
	mapAct = getMap()
	if !maps.Equal(mapAct, mapExp2) {
		t.Errorf("PutIf1 map %v exp %v", mapAct, mapExp2)
	}

	// PutIf false should not update
	reset()
	wasNewKey = rwmap.PutIf(key1, value2, putIfFalse)
	if wasNewKey {
		t.Error("PutIf2 wasNewKey true")
	}
	mapAct = getMap()
	if !maps.Equal(mapAct, mapExp1) {
		t.Errorf("PutIf2 map %v exp %v", mapAct, mapExp1)
	}

	// PutIf true should update
	reset()
	wasNewKey = rwmap.PutIf(key1, value2, putIfTrue)
	if wasNewKey {
		t.Error("PutIf2 wasNewKey true")
	}
	mapAct = getMap()
	if !maps.Equal(mapAct, mapExpValue2) {
		t.Errorf("PutIf2 map %v exp %v", mapAct, mapExpValue2)
	}
}

// ITEST= go test -race -v -run '^TestRWMapRace$' ./pmaps
func TestRWMapRace(t *testing.T) {
	randomLength := 16
	limitedSliceSize := 100
	lap := 100
	value := 3
	duration := time.Second

	// check environment
	if _, ok := os.LookupEnv("ITEST"); !ok {
		t.Skip("ITEST not present")
	}

	var limitedSlice = make([]string, limitedSliceSize)
	for i := 0; i < limitedSliceSize; i++ {
		limitedSlice[i] = randomAZ(randomLength)
	}

	var rwMap RWMap[string, int]
	NewRWMap2[string, int](&rwMap)
	var ctx, cancelFunc = context.WithCancel(context.Background())
	defer cancelFunc()
	//rand.Seed(time.Now().UnixNano())

	// put thread
	go func() {
		for ctx.Err() == nil {
			for _, randomString := range limitedSlice {
				rwMap.Put(randomString, value)
			}
			for i := 0; i < lap; i++ {
				rwMap.Put(randomAZ(randomLength), value)
			}
		}
	}()

	// get thread
	go func() {
		for ctx.Err() == nil {
			for _, randomString := range limitedSlice {
				rwMap.Get(randomString)
			}
			for i := 0; i < lap; i++ {
				rwMap.Get(randomAZ(randomLength))
			}
		}
	}()

	time.Sleep(duration)
}

// randomAZ provides a string of random characters using base64 encoding
//   - characters: a-zA-Z0-9+/
//   - use rand.Seed for randomization
func randomAZ(length int) (s string) {
	if length < 1 {
		return
	}
	// base64 encodes 64 values per character, ie. 6/8 bits as in 3 bytes into 4 bytes
	// 1 random byte provides 4/3 characters, ie factor 3/4, and add 1 due to integer truncation
	p := make([]byte, (length+1)*3/4)
	rand.Read(p)
	s = base64.StdEncoding.EncodeToString(p)
	if len(s) > length {
		s = s[:length]
	}
	return
}

// mapRanger ranges a map Range method
type mapRanger[K comparable, V any] struct {
	M         map[K]V
	keepGoing bool
}

// newMapRanger returns a tester for a map’s Range method
func newMapRanger[K comparable, V any](keepGoing bool) (ranger *mapRanger[K, V]) {
	return &mapRanger[K, V]{M: make(map[K]V), keepGoing: keepGoing}
}

// rangeFunc can be provided to a map’s Range method
func (m *mapRanger[K, V]) rangeFunc(key K, value V) (keepGoing bool) {
	m.M[key] = value
	return m.keepGoing
}
