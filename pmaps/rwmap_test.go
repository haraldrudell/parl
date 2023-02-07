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
	"testing"
	"time"

	"github.com/haraldrudell/parl/parli"
	"golang.org/x/exp/slices"
)

func TestRWMap(t *testing.T) {
	k1 := "key1"
	v1 := 1
	k2 := "key2"
	v2 := 2
	exp := []int{v1, v2}
	exp2 := []int{v2}

	var m parli.ThreadSafeMap[string, int]
	var m2 parli.ThreadSafeMap[string, int]
	var list []int
	var list2 []int
	newV := func() (value *int) { return &v1 }
	makeV := func() (value int) { return v2 }

	m = NewRWMap[string, int]()

	m.GetOrCreate(k1, newV, nil)
	m.GetOrCreate(k1, nil, nil)
	m.GetOrCreate(k2, nil, makeV)
	list = m.List()
	slices.Sort(list)
	if slices.Compare(list, exp) != 0 {
		t.Errorf("bad list %v exp %v", list, exp)
	}

	m.Delete(k1)
	m.Put(k2, v2)
	list = m.List()
	slices.Sort(list)
	if slices.Compare(list, exp2) != 0 {
		t.Errorf("bad list2 %v exp %v", list, exp2)
	}

	m2 = m.Clone()
	list2 = m2.List()
	slices.Sort(list2)
	if slices.Compare(list, list2) != 0 {
		t.Errorf("bad clone %v exp %v", list2, list)
	}

	m.Clear()
	if m.Length() != 0 {
		t.Errorf("bad length %d exp %d", m.Length(), 0)

	}

	(NewRWMap[string, int]()).GetOrCreate(k1, nil, nil)
	(NewRWMap[string, int]()).Get(k1)
	(NewRWMap[string, int]()).List()
	(NewRWMap[string, int]()).Put(k1, v1)
	(NewRWMap[string, int]()).Delete(k1)
	(NewRWMap[string, int]()).Clear()
	(NewRWMap[string, int]()).Length()
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

	var rwMap RWMap[string, int] = *NewRWMap2[string, int]()
	var ctx, cancelFunc = context.WithCancel(context.Background())
	defer cancelFunc()
	rand.Seed(time.Now().UnixNano())

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
