/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// WaitGroup1 is a [sync.WaitGroup] that starts at value 1
//   - purpose is lazy-initialization
//   - [WaitGroup1.IsDone] checks for value available
//   - [WaitGroup1.AddWin] picks winner
//   - winner resolves using [WaitGroup1.Done]
//   - —
//   - initialization-free lock-free thread-safe
//   - alternatives are:
//     [sync.Once] 20.27 ns: sync.Mutex wrapped by atomic.Uint32
//     [sync.Mutex] 14 ns per thread
//     [atomic.Pointer] costs speculative allocations
//     atomic thread-counter: very slow
//   - single-thread: 7.8% 1.39 ns slower than sync.Once go1.24.2 (17.63-16.24)/17.63
//   - parallel: 16.7% 111.5 ns faster than sync.Once (667.5-556.0)/667.5
//   - cpu profiling 250518: 17.08 ns.
//     initialize including winner-selection: 7 ns.
//     Done: 8 ns. IsDone 0.6 ns. Other 1.5 ns
//
// Usage:
//
//	var lazyX X
//	var w parl.WaitGroup1
//	if !w.IsDone() && w.AddWin() {
//	  lazyX = MakeX()
//	  w.Done()
//	}
//	// lazyX is thread-safe readable singleton
type WaitGroup1 struct {
	// Forces 64-bit alignment
	_ [0]uint64
	// wg is the waitgroup
	//	- methods: Add() Done() Wait()
	//	- aligned to 64-bit
	wg sync.WaitGroup
	// isReady is true if wg was initialized to one
	//	- separate 1 ns write+read atomic
	isReady atomic.Uint32
	// isWgWinner select thread to initialize wg
	//	- 0.6272 ns Load + 5.763 ns Swap
	isWgWinner atomic.Uint32
	// isDone is true if [WaitGroup1.Done] was invoked
	//	- separate 1 ns write+read atomic
	isDone atomic.Uint32
}

// AddWin initializes to one and returns whether this is winner
//   - isWinner true: this was the first AddWin
//   - losersWait NoOnceWait: loser threads do not await winner thread
//   - —
//   - upon return, it is certain the wait-group had counter 1
//   - similar to Add(1)
//
// Usage:
//
//	var lazyX X
//	var w parl.WaitGroup1
//	if !w.IsDone() && w.AddWin() {
//	  lazyX = MakeX()
//	  w.Done()
//	}
//	// lazyX is thread-safe readable singleton
func (w *WaitGroup1) AddWin(losersWait ...OnceChStrategy) (
	isWinner bool) {

	// consumers are responsible for synchronizing AddWin and Done
	//	- Done panics on isReady false
	//	- There is no Wait or Add methods modifying counter or causing panic
	//	- initialization must be ensured prior to isReady true
	if w.isReady.Load() == u32False {
		isWinner = w.initialize()
	}

	if isWinner ||
		(len(losersWait) > 0 && losersWait[0] == NoOnceWait) {
		return
	}
	w.wg.Wait()

	return
}

// IsDone returns true if Done was invoked
func (w *WaitGroup1) IsDone() (isDone bool) { return w.isDone.Load() == u32True }

// Done decrements the counter by one
//   - similar to [sync.Waitgroup.Done] with preceding initialization
func (w *WaitGroup1) Done() {

	// set isDone true: <= 1 ns
	if w.isDone.Load() == u32False {
		w.isDone.Store(u32True)
	}

	// invoke Done
	w.wg.Done()
}

// initialize ensures wg to be initialized and picks winner thread
//   - isReady bit must be false on invocation
//   - isWinner: true if this thread initialized wg
//   - —
//   - because the wg counter can be changed by Done,
//     counter zero-to-one CAS by all threads
//     cannot be used and is anyway too expensive at 20 ns
//   - any state-value is therefore separate from wg,
//     creating a synchronization need
//   - no thread can be allowed to return prior to
//     wg initialization completes
//   - therefore, subsequent threads require wait-mechanic.
//     The fastest wait mechanic is wait-group
func (w *WaitGroup1) initialize() (isWinner bool) {

	// winner: encountered value zero, and had swap return zero, too
	//	- — the Load is 0.6272 ns, the once or rare Swap is 5.763 ns
	//	- subsequent threads advance 5 ns faster than the winner thread
	if w.isWgWinner.Load() == u32False && w.isWgWinner.Swap(u32True) == u32False {

		isWinner = true
		// 0.3120 ns: save 6.7 ns
		var u64p = (*atomic.Uint64)(unsafe.Pointer(&w.wg))
		u64p.Store(deltaOne)
		// 0.3120 ns
		w.isReady.Store(u32True)
		return // initialized return: isWinner true
	}
	// this thread must wait for wg initialization complete
	//	- only additional threads arriving within 7.6 ns of first thread
	//	- typically none at all

	// spin-lock up to 5.6 ns
	for w.isReady.Load() == u32False {
		var uint64 = 0
		_ = uint64
	}

	return // wait complete return: isWinner false
}

const (
	// deltaOne is waitgroup state value for counter one
	deltaOne = uint64(1 << 32)
	// u32False is zero-value for uint32
	u32False = uint32(0)
	// u32True is non-zero-value for uint32
	u32True = uint32(1)
)
