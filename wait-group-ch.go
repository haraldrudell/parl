/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
)

// WaitGroupCh is [sync.WaitGroup] with channel-wait mechanic.
// WaitGroupCh is wait-free-locked and inspectable.
// WaitGroupCh waits for a collection of goroutines to finish.
//   - [WaitGroupCh.Add]: The consumer increments
//     the counter for each created goroutine
//   - — can also use [WaitGroupCh.Counts]
//   - [WaitGroupCh.Done] : goroutines decrements the counter as they exit
//   - — can also use [WaitGroupCh.DoneBool] [WaitGroupCh.Counts] [WaitGroupCh.Add]
//   - [WaitGroupCh.Wait] or [WaitGroupCh.Ch] is used to block until all goroutines have exited
//   - A WaitGroup must not be copied after first use.
//     In the terminology of the Go memory model, decrementing
//     “synchronizes before” the return of any Wait call or Ch read that it unblocks.
//   - panic: negative counter from invoking decreasing methods is panic
//   - panic: adjusting away from zero after invocation of [WaitGroupCh.Ch] or
//     [WaitGroupCh.Wait] is panic
//     NOTE: all Add should happen prior to invoking Ch or Wait
//   - —
//   - WaitGroupCh method-set is a superset of [sync.WaitGroup]
//   - — use of interfaces [WaitLegacy] and [DoneLegacy] makes replaceable
//   - WaitGroupCh is wait-free, observable, initialization-free, thread-safe
//   - channel-wait mechanic allows the consumer to be wait-free-locked
//     Progress by the consumer-thread is not prevented since:
//   - — the channel can be read non-blocking
//   - — consumers can wait for multiple channel events
//   - — consumers are not contending for a lock with any other thread
//   - there are race conditions:
//   - — writing a zero-counter record with unclosed channel,
//     therefore closing the channel
//   - — reading channel event while a zero-counter record exists,
//     therefore closing the channel
//   - — impact is that a panic might be missed
//   - —
//
// Usage:
//
//	var w parl.WaitGroupCh
//	w.Add(1)
//	go someFunc(&w)
//	…
//	<-w.Ch()
//	func someFunc(w parl.Doneable) {
//	  defer w.Done()
type WaitGroupCh struct {
	//	- awaitable lazy-initialized
	ch Awaitable
	// combined 32-bit adds and 32-bit dones
	//	- max 4 billion of each
	addsDones atomic.Uint64
	// set to true on [WaitGroupCh.Wait] [WaitGroupCh.Ch]
	isWaiting atomic.Bool
}

// WaitGroupCh implements Waiter
var _ Waiter = &WaitGroupCh{}

// WaitGroupCh is [DoneLegacy], ie. [sync.WaitGroup] compatible
var _ DoneLegacy = &WaitGroupCh{}

// WaitGroupCh is [Waitable], ie. [sync.WaitGroup] compatible
var _ Waitable = &WaitGroupCh{}

// DoneBool decrements the WaitGroup counter by one
//   - isExit true: the counter reached zero
//   - DoneBool when counter already zero is panic
func (w *WaitGroupCh) DoneBool() (isExit bool) {
	var count, _ = w.Counts(-1)
	isExit = count == 0
	return
}

// Count returns the current number of remaining threads
func (w *WaitGroupCh) Count() (currentCount int) {
	currentCount, _ = w.Counts()
	return
}

// Counts returns the current state optionally adjusting the counter
//   - delta: optional counter adjustment
//   - — delta negative and larger than current count is panic
//   - — delta present after the counter reduced to zero is panic
//   - currentCount: remaining count after any delta applied
//   - totalAdds: cumulative positive adds over WaitGroup lifetime after any delta applied
func (w *WaitGroupCh) Counts(delta ...int) (currentCount, totalAdds int) {

	// aggregated done count read from addsDone
	var dones int
	// atomic raw value
	var oldAddsDones uint64
	// read current adds and dones values
	totalAdds, dones, oldAddsDones = w.getAddsDones()
	// d is the delta to use
	var d int
	if len(delta) > 0 {
		d = delta[0]
	}

	// delta absent or zero case: return state
	if d == 0 {
		currentCount = totalAdds - dones
		return // no adjustment return
	}
	// adds is to be updated with d

	// check delta value
	var isPositive = d >= 0
	var positiveDelta uint64
	if isPositive {
		positiveDelta = uint64(d)
	} else {
		positiveDelta = uint64(-d)
	}
	if positiveDelta >= math.MaxUint32 {
		panic(perrors.ErrorfPF("delta too large: %d max ±%d",
			d, math.MaxUint32,
		))
	}
	// isZero is true if counter is zero
	var isZero bool

	// atomic update loop
	//	- values are in totalAdds dones isPositive positiveDelta
	for {

		isZero = totalAdds == dones
		if isZero && w.isWaiting.Load() {
			panic(perrors.ErrorfPF("attempt to adjust counter away from zero after Ch or Wait invoked: delta: %d adds: %d",
				d, totalAdds,
			))
		}

		// values for atomic
		var addsUint32, donesUint32 uint32

		// apply delta to local variables
		if isPositive {
			// update adds to a larger number case
			var addsUint64 = uint64(totalAdds) + positiveDelta
			if addsUint64 > math.MaxUint32 {
				panic(perrors.ErrorfPF("positive delta too large: %d + %d > %d",
					d, totalAdds, math.MaxUint32,
				))
			}
			addsUint32 = uint32(addsUint64)
			donesUint32 = uint32(dones)
		} else {
			// update dones to a larger number
			var donesUint64 = uint64(dones) + positiveDelta
			if donesUint64 > math.MaxUint32 {
				panic(perrors.ErrorfPF("negative delta too large: d: %d dones: %d > %d",
					d, dones, math.MaxUint32,
				))
			}
			addsUint32 = uint32(totalAdds)
			donesUint32 = uint32(donesUint64)
		}

		// attempt to write new values
		var newAddsDones = w.makeAddsDones(addsUint32, donesUint32)
		if !w.addsDones.CompareAndSwap(oldAddsDones, newAddsDones) {
			// another thread updated atomics

			// re-read current adds and dones values
			totalAdds, dones, oldAddsDones = w.getAddsDones()
			continue // retry atomic update
		}
		isZero = addsUint32 == donesUint32
		totalAdds = int(addsUint32)
		currentCount = totalAdds - int(donesUint32)
		break
	}
	// addsDones was updated

	// check channel should close
	if isZero {
		if !w.ch.Close(EventuallyConsistency) {
			panic(perrors.NewPF("multiple Counts invocations reached zero"))
		}
	}

	return
}

// IsZero returns whether the counter is currently zero
func (w *WaitGroupCh) IsZero() (isZero bool) {
	var adds, dones, _ = /*addsDones*/ w.getAddsDones()
	isZero = adds == dones
	return
}

// Add adds delta, which may be negative, to the WaitGroup counter
//   - If the counter becomes zero, all goroutines blocked on Wait are released
//   - If the counter goes negative, Add panics
func (w *WaitGroupCh) Add(delta int) { w.Counts(delta) }

// Done decrements the WaitGroup counter by one.
func (w *WaitGroupCh) Done() { w.Counts(-1) }

// Ch returns a channel that closes once the counter reaches zero
func (w *WaitGroupCh) Ch() (awaitableCh AwaitableCh) {
	if !w.isWaiting.Load() {
		w.isWaiting.Store(true)
	}
	awaitableCh = w.ch.Ch()

	return
}

// Wait blocks until the WaitGroup counter is zero
//   - wait-free-locked via [WaitGroupCh.Ch]
func (w *WaitGroupCh) Wait() { <-w.ch.Ch() }

// Reset triggers the current channel and resets the WaitGroup
func (w *WaitGroupCh) Reset() (w2 *WaitGroupCh) {
	w2 = w
	w.ch.Close()
	w.ch = Awaitable{}
	w.addsDones.Store(0)
	w.isWaiting.Store(false)

	return
}

// getAddsDones reads the combined uint64 atomic value
func (w *WaitGroupCh) getAddsDones() (adds, dones int, addsDones uint64) {
	// uint64
	addsDones = w.addsDones.Load()
	// dones is higher 32 bits
	dones = int(addsDones >> 32)
	// adds in lower 32 bits
	adds = int(addsDones & math.MaxUint32)

	return
}

// getAddsDones reads the combined uint64 atomic value
func (w *WaitGroupCh) makeAddsDones(adds, dones uint32) (addsDones uint64) {
	return uint64(adds) | uint64(dones)<<32
}

func (w *WaitGroupCh) String() (s string) {
	var adds, dones, _ = /*addsDones*/ w.getAddsDones()
	var count = adds - dones
	s = fmt.Sprintf("waitGroupCh_count:%d(adds:%d)_isWaiting:%t_isClosed:%t",
		count, adds, w.isWaiting.Load(), w.ch.IsClosed(),
	)
	return
}

// Add() Done() Wait()
var _ sync.WaitGroup

// func (*sync.WaitGroup).Add(delta int)
var _ = (&sync.WaitGroup{}).Add

// func (*sync.WaitGroup).Wait()
var _ = (&sync.WaitGroup{}).Wait

// func (*sync.WaitGroup).Done()
var _ = (&sync.WaitGroup{}).Done
