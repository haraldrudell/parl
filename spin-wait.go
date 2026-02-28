/*
© 2026–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"runtime"
	"sync/atomic"

	"golang.org/x/sys/cpu"
)

const (
	ResetExponentialNo ResetExponential = iota
	NoExponential
)

type ResetExponential uint8

// SpinWait provides a processor low-power wait function
// exponentially 70–600 ns on 2021 Apple M
//
// Why:
//   - provides a waiting function to be used for less than 10 µs waits where
//     [sync.Mutex] 150 ms thread-suspends are unfavorable
//   - may exclude 168 µs [runtime.Gosched] that shares the execution unit with
//     other goroutines
//   - wait may be exponential for additional performance or using a constant
//     wait time for occasional waits
//   - may include a safe-point allowing for stop-the-world Go garbage
//     collection. Execution without safe-points may cause dead-lock
//     among goroutines
//   - written 260227 by Harald Rudell
//
// Usage:
//
//	var spinWait parl.SpinWait
//	…
//	spinWait.ResetExponential()
//	for {
//	  if isOK() {
//	    break
//	  }
//	  spinWait.Wait()
//	}
type SpinWait struct {
	MaxSpinCount int
	DoGosched    bool
	DoSafePoint  bool
	_            cpu.CacheLinePad
	spinCount    atomic.Uint32
	_            cpu.CacheLinePad
}

// ResetExponential …. Thread-safe
func (w *SpinWait) ResetExponential() {

	// attain inter-thread write visibility for this thread
	//	- applies to other SpinWait fields
	w.spinCount.Load()

	// reset spin-count
	w.spinCount.Store(1)
}

// Wait …. Thread-safe
func (w *SpinWait) Wait(resetExponential ResetExponential) {

	// attain inter-thread write visibility
	var spinCount = w.spinCount.Load()

	// ensure initialized
	if w.MaxSpinCount == 0 {
		w.MaxSpinCount = maxSpinCount
		w.DoGosched = true
		w.DoSafePoint = true
		spinCount = 1
		w.spinCount.Store(spinCount)
	}

	// do any [runtime.Gosched]
	var doGosched = w.DoGosched
	if !doGosched && goExecutionUnits == 1 {
		doGosched = true
	}
	if doGosched && !runtime_canSpin(1) {
		runtime.Gosched()
	}

	// do exponential delay
	for i := range spinCount {
		_ = i
		if w.DoSafePoint {
			safePoint()
		}
		runtime_doSpin()
	}

	// exponential wait
	if resetExponential == NoExponential ||
		int(spinCount) >= w.MaxSpinCount {
		return
	}
	spinCount <<= 1
	for {
		var old = w.spinCount.Load()
		if old >= spinCount || w.spinCount.CompareAndSwap(old, spinCount) {
			break
		}
	}
}

var (
	goExecutionUnits = runtime.GOMAXPROCS(0)
)
