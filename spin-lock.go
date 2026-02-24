/*
© 2026–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"runtime"
	"sync/atomic"

	_ "unsafe"
)

// SpinLock is a lock that does not suspend threads
//
// Why:
//   - non-suspend of threads is the unique value proposition
//   - — if maximum performance is required from the critical section
//     at the cost of high cpu, spin-lock prevents the all-thread suspend of Mutex
//   - — in particular, a worker thread receiving work items may suffer from
//     excessive suspend with Mutex
//   - a fast lock, parallel ≈1.06 µs BenchmarkUnboundedQueueAdd
//   - does not suspend threads like sync.Mutex
//   - — the problem with Mutex is a long thread wake-up-latency 165 ms BenchmarkUnblock
//   - compared to atomic access, lock provides when another thread’s operations has concluded
//   - atomic invocation-counters are too slow due to [atomic.Add] 646.5 ns BenchmarkAddP
//   - successful parallel [atomic.CompareAndSwap] is 2.285 µs BenchmarkCASP
//   - use of [atomic.Pointer] requires allocation on each write, maybe 10 µs
//   - in 2026 with 10 exeuction units, lock is the fastest
//
// Notes:
//   - if wait exceeds 10 µs, SpinLock causes high cpu
//   - sync.Mutex spins for 100 ns before suspend
//   - — if a thread holding the lock is suspended as threads are every 10 ms,
//     the tread may remain suspended for 150 ms.
//     Mutex then suspends all holding threads, too.
//     Mutex is designed to be better than spin-lock
//   - computer science has max recommended critical-section latency
//     two context switch overhead: 10 µs
//   - spin-lock latency ≈ work×number of contenders, ie. holding goroutines
//   - critical section latency 100 ns, 10 goroutines: 100 ns × 10 ≈ 1 µs
//   - — the execution unit count is not the number to use
//
// Design:
//   - exponential back-off up to 600 ns between lock-attempts
//   - invokes runtime_canSpin for Go stop-the-world garbage collection
//   - core delay-mechanic runtime_doSpin 68.7 ns
//   - initialization-free, functional chaining, deferrable
//   - written 260218 by Harald Rudell
type SpinLock struct {
	// lock atomic is the lock
	//	- values: spinUnlocked spinLocked
	lock atomic.Uint32
}

// IsHeld returns true if the lock is currently held
func (s *SpinLock) IsHeld() (isHeld bool) { return s.lock.Load() == spinLocked }

// Lock acquires the lock blocking
//
// Usage:
//
//	defer lock.Lock().Unlock()
func (s *SpinLock) Lock() (m2 Unlocker) {
	m2 = s

	// channel has a loop checking the atomic once per loop
	//	- in the loop, channel invokes procyield(30), 6–10 ns
	//	- after 4 loops the thread is suspended
	//	- total duration has been said to be 120–150 ns

	// spinCount is number of loops
	// for exponential delay
	var spinCount = 1

	// adding Load prior to CompareAndSwap makes
	// benchmark go to 220.4 wall-ns/op from 243.1
	// BenchmarkSpinLock 260218
	for s.lock.Load() == spinLocked || !s.lock.CompareAndSwap(spinUnlocked, spinLocked) {

		// benchmarking with BenchmarkSpinLock 260218
		//	- range 30: benchmark result: 243.1 ns, loop delay: is 2.061 μs
		//	- range 300: 103.4 ns
		//	- range 3000: 89.27 ns
		//	- range 30_000 88.25 ns
		//	- 3000 runtime_doSpin is 206.083 μs, each is 68.7 ns
		//	- reality is that huge delay reduces the number of contending threads and
		//		sabotages the test with invalid but great numbers
		//	- instead tune for benchmark of the UnboundedQueue consumer
		//	- runtime_doSpin should be invoked once or twice between reading atomic

		// BenchmarkUnboundedQueueAdd 260218 with runtime.GC in benchmark
		//	- 1 doSpin + 1 canSpin: 5.140 μs
		//	- 2 doSpin + 1 canSpin: 4.798 μs
		//	- 0 doSpin + 1 canSpin: 6.069 μs
		//	- 3 doSpin + 1 canSpin: 3.632 μs
		//	- the design is one thread at a time can do work inside the lock
		//	- as spin-time increases with 10 threads at the lock,
		//		the number of threads that can do work
		//		does not increase.
		//	- exponential back-off up to 8, 2× increase per lap: 1.978 μs
		//	- time inside the outer lock is about 100 ns
		//	- time outside the outer lock is about 4 µs
		//	- time inside the inner lock is about 25 ns
		//	- at 8, loop delay is about 562 ns — 8*68.7+12.5
		//	- spin-lock latency ≈ work×number of contenders
		//	- 100 ns × 10 ≈ 1 µs
		//	- with exponential back-off benchmark result is 1.913 µs

		// P-M-G
		//	- P is logical proocessor context 256 slots. Go has GoMaxProcs of them
		//	- M is the OS thread. OS assigns an execution unit to M
		//	- G is the goroutine. Go selects M

		//	- Mutex suspend releases M which takes 150 ms to resume
		//	- runtime.Gosched 168 μs releases G the goroutine
		//	- — M and P are still present and execute another goroutine
		//	- safepoints, eg. function calls, allows stop-the-world gv to run
		//	- GC releases P M and parks G
		//	- GC uses its goroutine to complete GC
		//	- other pauses may happen:
		//	- Go pauses gioroutines after 10 ms
		//	- OS suspend after time-slice expiration 4–24 ms

		// can spin returns false when other goroutines should run
		// 	- latency of runtime_canSpin is 12.5 ns
		if !runtime_canSpin(1) {
			// This allows the GC to finish its STW phase
			//	- typical latency 168 μs
			// - may happen every few seconds max 2 minutes
			runtime.Gosched()
		}

		// runtime_doSpin latency: 68.7 ns
		//	- not true: spin for 6–10 ns
		for i := range spinCount {
			_ = i

			// allow stop-the-world GC to intercept
			safePoint()

			runtime_doSpin()
		}
		if spinCount < maxSpinCount {
			spinCount <<= 1
		}
	}

	return
}

// Unlock releases the held lock
func (s *SpinLock) Unlock() { s.lock.Store(spinUnlocked) }

func safePoint() { safePoint2() }

//go:noinline
func safePoint2() {}

// runtime_doSpin
//   - command-click on GOARCH to display file from runtime package
//   - navigate to file src/runtime/proc.go
//   - runtime_doSpin delegates to procyield(30)
//   - delegates to procyieldAsm(cycles uint32)
var _ = runtime.GOARCH

const (
	// spinUnlocked is value when unlocked
	spinUnlocked = 0
	// spinLocked is value when locked
	spinLocked = 1
	// maximum number of loops for exponential back-off
	maxSpinCount = 8
)

// runtime_doSpin spins for 6–10 ns depending on processor speed 2026
//
//go:linkname runtime_doSpin sync.runtime_doSpin
func runtime_doSpin()

//go:linkname runtime_canSpin sync.runtime_canSpin
func runtime_canSpin(i int) bool
