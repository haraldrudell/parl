/*
© 2024–present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import (
	"reflect"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

// Open:closed/Open Ch Close:guaranteed/eventual IsClosed:open/closed
var cyclicAwaitablesuite = []func(b *testing.B){
	BenchmarkCaCh,
	BenchmarkCaIsClosedTrue,
	BenchmarkCaIsClosedFalse,
	BenchmarkCaEventuallyConsistentTrue,
	BenchmarkCaEventuallyConsistentFalse,
	BenchmarkCaOpenOpened,
	BenchmarkCaOpenClosed,
}

// Running tool: /opt/homebrew/bin/go test -benchmem -run=^$ -bench ^BenchmarkCyclicAwitable$ github.com/haraldrudell/parl
//
// 240517 c66
// pkg: github.com/haraldrudell/parlBenchmarkCyclicAwitable/BenchmarkCaCh-10         	 4066071	       296.8 ns/op	         2.968 wall-ns/op	       0 B/op	       0 allocs/op
// BenchmarkCyclicAwitable/BenchmarkCaIsClosedTrue-10         	 4500962	       267.4 ns/op	         2.674 wall-ns/op	       0 B/op	       0 allocs/op
// BenchmarkCyclicAwitable/BenchmarkCaIsClosedFalse-10        	 1608205	       746.2 ns/op	         7.462 wall-ns/op	       0 B/op	       0 allocs/op
// BenchmarkCyclicAwitable/BenchmarkCaEventuallyConsistentTrue/BenchmarkEventuallyConsistentTrue-10         	550551525	         1.826 ns/op	        18.26 wall-ns/op	       0 B/op	       0 allocs/op
// BenchmarkCyclicAwitable/BenchmarkCaEventuallyConsistentFalse/BenchmarkEventuallyConsistentFalse-10       	631737799	         1.718 ns/op	        17.18 wall-ns/op	       0 B/op	       0 allocs/op
// BenchmarkCyclicAwitable/BenchmarkCaOpenOpened-10                                                         	 1474467	       815.4 ns/op	         8.154 wall-ns/op	       0 B/op	       0 allocs/op
// BenchmarkCyclicAwitable/BenchmarkCaOpenClosed-10                                                         	 1228254	       976.9 ns/op	         9.769 wall-ns/op	       0 B/op	       0 allocs/op
func BenchmarkCyclicAwitable(b *testing.B) {
	for _, bm := range cyclicAwaitablesuite {
		var cL, err = FuncName(bm)
		if err != nil {
			b.Fatalf("reflect failed: %s", perrors.Short(err))
		}
		b.Run(cL.FuncIdentifier(), bm)
	}
}

// 2.918 ns
//
// 240517 c66
// Running tool: /opt/homebrew/bin/go test -benchmem -run=^$ -bench ^BenchmarkCaCh$ codeberg.org/haraldrudell/goprogramming/pkg/github.com_haraldrudell_parl
// pkg: codeberg.org/haraldrudell/goprogramming/pkg/github.com_haraldrudell_parl
// BenchmarkCaCh-10    	 3926060	       291.8 ns/op	         2.918 wall-ns/op	       0 B/op	       0 allocs/op
func BenchmarkCaCh(b *testing.B) {

	var a CyclicAwaitable
	var batch = 100
	for i := 0; i < b.N; i++ {
		for i := 0; i < batch; i++ {
			a.Ch()
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(b.Elapsed())/float64(b.N)/float64(batch), "wall-ns/op")
}

// [CyclicAwaitable.IsClosed when closed: 2.620 ns
//
// 240517 c66
// Running tool: /opt/homebrew/bin/go test -benchmem -run=^$ -bench ^BenchmarkCaIsClosedTrue$ codeberg.org/haraldrudell/goprogramming/pkg/github.com_haraldrudell_parl
// pkg: codeberg.org/haraldrudell/goprogramming/pkg/github.com_haraldrudell_parl
// BenchmarkCaIsClosedTrue-10    	 4430820	       262.0 ns/op	         2.620 wall-ns/op	       0 B/op	       0 allocs/op
func BenchmarkCaIsClosedTrue(b *testing.B) {
	var a CyclicAwaitable
	a.Close()
	var batch = 100
	for i := 0; i < b.N; i++ {
		for i := 0; i < batch; i++ {
			a.IsClosed()
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(b.Elapsed())/float64(b.N)/float64(batch), "wall-ns/op")
}

// [CyclicAwaitable.IsClosed] when open: 6.191 ns
//
// 240517 c66
// Running tool: /opt/homebrew/bin/go test -benchmem -run=^$ -bench ^BenchmarkCaIsClosedFalse$ codeberg.org/haraldrudell/goprogramming/pkg/github.com_haraldrudell_parl
// pkg: codeberg.org/haraldrudell/goprogramming/pkg/github.com_haraldrudell_parl
// BenchmarkCaIsClosedFalse-10    	 1930954	       619.1 ns/op	         6.191 wall-ns/op	       0 B/op	       0 allocs/op
func BenchmarkCaIsClosedFalse(b *testing.B) {
	var a CyclicAwaitable
	var batch = 100
	for i := 0; i < b.N; i++ {
		for i := 0; i < batch; i++ {
			a.IsClosed()
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(b.Elapsed())/float64(b.N)/float64(batch), "wall-ns/op")
}

// 16.12 ns per op
//
// 240517 c66
// Running tool: /opt/homebrew/bin/go test -benchmem -run=^$ -bench ^BenchmarkCaEventuallyConsistentFalse$ codeberg.org/haraldrudell/goprogramming/pkg/github.com_haraldrudell_parl
// pkg: codeberg.org/haraldrudell/goprogramming/pkg/github.com_haraldrudell_parl
// BenchmarkCaEventuallyConsistentFalse/BenchmarkEventuallyConsistentFalse-10         	826047098	         1.612 ns/op	        16.12 wall-ns/op	       0 B/op	       0 allocs/op
func BenchmarkCaEventuallyConsistentFalse(b *testing.B) {
	var ec EventuallyConsistent
	b.Run("BenchmarkEventuallyConsistentFalse", NewCaCloseTest(ec).Benchmark)
}

// 16.99 ns
//
// 240517 c66
// Running tool: /opt/homebrew/bin/go test -benchmem -run=^$ -bench ^BenchmarkCaEventuallyConsistentTrue$ codeberg.org/haraldrudell/goprogramming/pkg/github.com_haraldrudell_parl
// pkg: codeberg.org/haraldrudell/goprogramming/pkg/github.com_haraldrudell_parl
// BenchmarkCaEventuallyConsistentTrue/BenchmarkEventuallyConsistentTrue-10         	660889242	         1.699 ns/op	        16.99 wall-ns/op	       0 B/op	       0 allocs/op
func BenchmarkCaEventuallyConsistentTrue(b *testing.B) {
	b.Run("BenchmarkEventuallyConsistentTrue", NewCaCloseTest(EventuallyConsistency).Benchmark)
}

// CaCloseTest tests Awaitable eventually consistent Close
type CaCloseTest struct{ evCon EventuallyConsistent }

// NewAwCloseTest returns a subbenchmark testing Awaitable eventually consistent Close
func NewCaCloseTest(evCon EventuallyConsistent) (a *CaCloseTest) { return &CaCloseTest{evCon: evCon} }

// subbenchmark testing Awaitable eventually consistent Close
func (t *CaCloseTest) Benchmark(b *testing.B) {
	var i Atomic64[int]
	var p0 atomic.Pointer[CyclicAwaitable]
	var a1 CyclicAwaitable
	p0.Store(&a1)
	b.ResetTimer()
	b.RunParallel(func(b *testing.PB) {
		var isFirstThread = i.Add(1) == 1
		var p = &p0
		var a *CyclicAwaitable
		var evConArg = t.evCon
		for b.Next() {
			a = p.Load()
			a.Close(evConArg)
			if isFirstThread {
				var a2 CyclicAwaitable
				p.Store(&a2)
			}
		}
	})
	b.StopTimer()
	b.ReportMetric(float64(b.Elapsed())/float64(b.N)*float64(runtime.GOMAXPROCS(0)), "wall-ns/op")
}

// 6.833 ns
//
// 240517 c66
// Running tool: /opt/homebrew/bin/go test -benchmem -run=^$ -bench ^BenchmarkCaOpenOpened$ codeberg.org/haraldrudell/goprogramming/pkg/github.com_haraldrudell_parl
// pkg: codeberg.org/haraldrudell/goprogramming/pkg/github.com_haraldrudell_parl
// BenchmarkCaOpenOpened-10    	 1739640	       683.3 ns/op	         6.833 wall-ns/op	       0 B/op	       0 allocs/op
func BenchmarkCaOpenOpened(b *testing.B) {
	var a CyclicAwaitable
	var batch = 100
	for i := 0; i < b.N; i++ {
		for i := 0; i < batch; i++ {
			a.Open()
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(b.Elapsed())/float64(b.N)/float64(batch), "wall-ns/op")
}

// 9.736 ns
//
// 240517 c66
// Running tool: /opt/homebrew/bin/go test -benchmem -run=^$ -bench ^BenchmarkCaOpenClosed$ github.com/haraldrudell/parl
// pkg: github.com/haraldrudell/parl
// BenchmarkCaOpenClosed-10    	 1233966	       973.6 ns/op	         9.736 wall-ns/op	       0 B/op	       0 allocs/op
func BenchmarkCaOpenClosed(b *testing.B) {
	var closedCyclic, ca CyclicAwaitable
	closedCyclic.Close()
	var batch = 100
	for i := 0; i < b.N; i++ {
		for i := 0; i < batch; i++ {
			ca.awp.Store(closedCyclic.awp.Load())
			closedCyclic.Open()
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(b.Elapsed())/float64(b.N)/float64(batch), "wall-ns/op")
}

// preflect.FuncName
func FuncName(funcOrMethod any) (cL *pruntime.CodeLocation, err error) {

	// funcOrMethod cannot be nil
	if funcOrMethod == nil {
		err = NilError("funcOrMethod")
		return
	}

	// funcOrMethod must be underlying type func
	var reflectValue = reflect.ValueOf(funcOrMethod)
	if reflectValue.Kind() != reflect.Func {
		err = perrors.ErrorfPF("funcOrMethod not func: %T", funcOrMethod)
		return
	}

	// get func name, “func1” for anonymous
	var runtimeFunc = runtime.FuncForPC(reflectValue.Pointer())
	if runtimeFunc == nil {
		err = perrors.NewPF("runtime.FuncForPC returned nil")
		return
	}
	cL = pruntime.CodeLocationFromFunc(runtimeFunc)

	return
}
