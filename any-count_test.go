/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"
	"testing"
	"unsafe"
)

func TestAnyCount(t *testing.T) {
	//t.Error("Logging on")
	const (
		expOne   = 1
		expThree = 3
	)
	var (
		values = []int{2, 3}
		total  int
	)

	// methods: Count() Init() Cond()
	var a = AnyCount[int]{Value: 1, HasValue: true}

	// Sizeof AnyCount[int]: 40
	t.Logf("Sizeof AnyCount[int]: %d", unsafe.Sizeof(a))

	// Count() 1
	if tot := a.Count(); tot != expOne {
		t.Errorf("Count %d exp %d", tot, expOne)
	}

	// iteration single value
	total = 0
	for v := range a.Seq {
		_ = v
		total++
	}
	if total != expOne {
		t.Errorf("Count %d exp %d", total, expOne)
	}

	// Count multiple values
	a.Values = values
	if tot := a.Count(); tot != expThree {
		t.Errorf("FAIL Count %d exp %d", tot, expThree)
	}

	// iteration multiple values
	total = 0
	for v := range a.Seq {
		_ = v
		total++
	}
	if total != expThree {
		t.Errorf("FAIL Count %d exp %d", total, expThree)
	}
}

// memory benchmark:
//   - appears in pprof as: github.com/haraldrudell/parl.BenchmarkAnyCount
//   - — there are no allocations. None. Zilch. Nada.
//
// F=tmp/anycount-$(hostname -s)-heap-$(date "+%FT%T%z").prof && go test -benchmem -memprofile=$F -memprofilerate=1 -run=^$ -bench ^BenchmarkAnyCount$ github.com/haraldrudell/parl && echo $F
func BenchmarkAnyCount(b *testing.B) {
	var (
		values = []int{2, 3}
	)

	// because AnyCount is allocation-free,
	// do an allocation here so that it appears in the heap profile
	//	- an atomic must be on the heap, ie. be allocated
	//	- the atomic must be used to not be optimized away
	//	- atomic.Bool causes 96 bytes allocation
	var a atomic.Bool
	a.Load()

	// here are some silly operations proving that no allocation occur
	for i := 0; i < b.N; i++ {

		// methods: Count() Init() Cond()
		var a = AnyCount[int]{Value: 1, HasValue: true}

		// Count() 1
		a.Count()

		// iteration single value
		for v := range a.Seq {
			_ = v
		}

		// Count multiple values
		a.Values = values
		a.Count()

		// iteration multiple values
		for v := range a.Seq {
			_ = v
		}
	}
}
