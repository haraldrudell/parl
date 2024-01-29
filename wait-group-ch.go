/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
)

// WaitGroupCh is like a [sync.WaitGroup] with channel-wait mechanic.
// therefore, unlike sync, WaitGroupCh is wait-free and observable.
// WaitGroupCh waits for a collection of goroutines to finish.
// The main goroutine increments the counter for each goroutine.
// Then each of the goroutines decrements the counter until zero.
// Wait or Ch can be used to block until all goroutines have finished.
// A WaitGroup must not be copied after first use.
// In the terminology of the Go memory model, decrementing
// “synchronizes before” the return of any Wait call or Ch read that it unblocks.
//   - counter is increased by [WaitGroupCh.Add] or [WaitGroupCh.Count]
//   - counter is decreased by [WaitGroupCh.Done] [WaitGroupCh.Add]
//     [WaitGroupCh.DoneBool] or [WaitGroupCh.Count]
//   - counter zero is awaited by [WaitGroupCh.Ch] or [WaitGroupCh.Wait]
//   - observability if provided by [WaitGroupCh.Count] [WaitGroupCh.DoneBool] and
//     [WaitGroupCh.IsZero]
//   - panic: negative counter from invoking decreasing methods is panic
//   - panic: adjusting away from zero after invocation of [WaitGroupCh.Ch] or
//     [WaitGroupCh.Wait] is panic.
//     NOTE: all Add should happen prior to invoking Ch or Wait
//   - —
//   - WaitGroupCh is wait-free, observable, initialization-free, thread-safe
//   - channel-wait mechanic allows the consumer to be wait-free
//     Progress by the consumer-thread is not prevented since:
//   - — the channel can be read non-blocking
//   - — consumers can wait for multiple channel events
//   - — consumers are not contending for a lock with any other thread
//   - WaitGroupCh method-set is a superset of [sync.WaitGroup]
//   - there are race conditions:
//   - — writing a zero-counter record with unclosed channel,
//     therefore closing the channel
//   - — reading channel event while a zero-counter record exists,
//     therefore closing the channel
//   - — impact is that a panic might be missed
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
	// p as atomic pointer provides integrity in reading counters without a lock
	//	- atomic Pointer enables initialization-free operation
	p atomic.Pointer[addsDones]
}

// addsDone provides integrity in reading counters without a lock
//   - thread-safe access is provided by reading WaitGroupCh.p
type addsDones struct {
	// pointer to be shared across generations
	ch *Awaitable
	// cumulative positive adds
	adds int
	// cumulative negative adds
	dones int
	// flags that further adjustments are not possible
	channelAboutToClose bool
}

// Done decrements the WaitGroup counter by one.
func (w *WaitGroupCh) DoneBool() (isExit bool) {
	var count, _ = w.Count(-1)
	return count == 0
}

// Count returns the current state optionally adjusting the counter
//   - delta is optional counter adjustment
//   - currentCount is current remaining count
//   - totalAdds is cumulative positive adds over WaitGroup lifetime
func (w *WaitGroupCh) Count(delta ...int) (currentCount, totalAdds int) {
	var newAddsDones addsDones
	var channelShouldClose bool
	for {
		var p = w.getP()

		// delta absent or zero case: return state
		if len(delta) == 0 || delta[0] == 0 {
			currentCount = p.adds - p.dones
			totalAdds = p.adds
			return // no adjustment return
		}

		// check for channel about to close state
		//	- d is known to be non-zero
		var d = delta[0]
		if p.channelAboutToClose {
			panic(perrors.ErrorfPF("attempt to adjust away from zero when Ch or Wait invoked: delta: %d", d))
		}

		// create new record
		newAddsDones = *p
		if d > 0 {
			// increment case
			newAddsDones.adds += d
		} else if d < 0 {
			// decrement case
			if p.dones+(-d) > p.adds {
				panic(perrors.ErrorfPF("attempt to adjust to negative: d: %d adds: %d dones: %d",
					d, p.adds, p.dones,
				))
			}
			newAddsDones.dones += (-d)
			// check channel should close
			if newAddsDones.dones == newAddsDones.adds {
				if channelShouldClose = !newAddsDones.channelAboutToClose; channelShouldClose {
					newAddsDones.channelAboutToClose = true
				}
			}
		}
		if w.p.CompareAndSwap(p, &newAddsDones) {
			break // successfully wrote new record
		}
	}
	currentCount = newAddsDones.adds - newAddsDones.dones
	totalAdds = newAddsDones.adds
	if !channelShouldClose {
		return // not closing channel now
	}

	// trig the awaitable
	newAddsDones.ch.Close()

	return
}

// IsZero returns whether the counter is currently zero
func (w *WaitGroupCh) IsZero() (isZero bool) {
	var p = w.getP()
	return p.adds == p.dones
}

// Add adds delta, which may be negative, to the WaitGroup counter
//   - If the counter becomes zero, all goroutines blocked on Wait are released
//   - If the counter goes negative, Add panics
func (w *WaitGroupCh) Add(delta int) { w.Count(delta) }

// Done decrements the WaitGroup counter by one.
func (w *WaitGroupCh) Done() { w.Count(-1) }

// Ch returns a channel that closes once the counter reaches zero
func (w *WaitGroupCh) Ch() (awaitableCh AwaitableCh) { return w.getCh() }

// Wait blocks until the WaitGroup counter is zero.
func (w *WaitGroupCh) Wait() { <-w.getCh() }

// Reset triggers the current channel and resets the WaitGroup
func (w *WaitGroupCh) Reset() (w2 *WaitGroupCh) {
	w2 = w
	var p = w.p.Swap(nil)
	if p == nil {
		return // was not initialized
	}
	p.ch.Close()

	return
}

// getCh returns the channel and notes the channel was read
func (w *WaitGroupCh) getCh() (awaitableCh AwaitableCh) {
	var p *addsDones
	var newAddsDones addsDones
	var shouldCloseChannel bool
	for {
		p = w.getP()
		if shouldCloseChannel = !p.channelAboutToClose && p.adds == p.dones; !shouldCloseChannel {
			break // no requirement to take channel close action
		}
		newAddsDones = *p
		p.channelAboutToClose = true
		if w.p.CompareAndSwap(p, &newAddsDones) {
			p.ch.Close()
			break // closed the channel
		}
	}

	return p.ch.Ch()
}

// get ensures that WaitGroupCh is initialized
func (w *WaitGroupCh) getP() (p *addsDones) {
	var newAddsDones *addsDones
	for {
		if p = w.p.Load(); p != nil {
			return // already initialized return
		} else if newAddsDones == nil {
			newAddsDones = &addsDones{ch: &Awaitable{}}
		}
		if w.p.CompareAndSwap(nil, newAddsDones) {
			p = newAddsDones
			return // initialized with new p
		}
	}
}

func (w *WaitGroupCh) String() (s string) {
	var p = w.getP()
	return fmt.Sprintf("waitGroupCh_count:%d(adds:%d)", p.adds-p.dones, p.adds)
}

// func (*sync.WaitGroup).Add(delta int)
// func (*sync.WaitGroup).Done()
// func (*sync.WaitGroup).Wait()
var _ sync.WaitGroup
var _ = (&sync.WaitGroup{}).Add
var _ = (&sync.WaitGroup{}).Wait
var _ = (&sync.WaitGroup{}).Done
