/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"
)

// PeriodWaiter temporarily holds threads invoking Wait
//   - HoldWaiters causes any Wait invokers to block until ReleaseWaiters
//
// that has not been succeeded by a ReleaseWaiters invocation. Thread-safe
type PeriodWaiter struct {
	// gen is initialized by HoldWaiters and set to nil by
	// ReleaseWaiters
	gen atomic.Pointer[periodWaiter]
}

type periodWaiter struct {
	// waitGroup holds Wait invokers
	wg WaitGroupCh
	// number of threads in Wait
	count atomic.Uint64
}

// NewPeriodWaiter returns an object that can hold threads temporarily. Thread-safe
//   - initially [PeriodWaiter.Wait] invokers are not held
//   - After invoking [PeriodWaiter.], Wait invokers are held until an invocation of
//     [PeriodWaiter.ReleaseWaiters]
func NewPeriodWaiter() (periodWaiter *PeriodWaiter) { return &PeriodWaiter{} }

// HoldWaiters causes a thread invoking [PeriodWaiter.Wait] to wait. Thread-safe
//   - idempotent
func (p *PeriodWaiter) HoldWaiters() {

	// check for already waiting
	if p.gen.Load() != nil {
		return // already waiting return: noop
	}

	// try to add a created wait group
	var pw periodWaiter
	// add one causing [WaitGroupCh.Wait] to block
	pw.wg.Add(1)
	// atomic nil to value of p.wg
	p.gen.CompareAndSwap(nil, &pw)
}

// ReleaseWaiters releases any threads blocked in [PeriodWaiter.Wait]
// and lets new Wait invokers proceed. Thread-safe
func (p *PeriodWaiter) ReleaseWaiters() {

	// atomic read of the current wait group
	var pw = p.gen.Load()
	if pw == nil {
		return // state not waiting return: noop
	}

	// resolve any successfully obtained waitgroup
	if !p.gen.CompareAndSwap(pw, nil) {
		return // other thread already niled it: noop
	}
	pw.wg.Done()
}

// Count returns the number of threads currently in Wait
//   - threads invoking [PeriodWaiter.Ch] or
func (p *PeriodWaiter) Count() (waitingThreads int) {
	if pw := p.gen.Load(); pw != nil {
		waitingThreads = int(pw.count.Load())
	}
	return
}

// IsHold returns true if Wait will currently block
func (p *PeriodWaiter) IsHold() (isHold bool) { return p.gen.Load() != nil }

// Ch returns a channel that closes on ReleaseWaiters
//   - ch is nil if not currently waiting
func (p *PeriodWaiter) Ch() (ch AwaitableCh) {
	if pw := p.gen.Load(); pw != nil {
		ch = pw.wg.Ch()
	}
	return
}

// Wait blocks the thread if a HoldWaiters invocation took place with no
// ReleaseWaiters succeeding it. Thread-safe
func (p *PeriodWaiter) Wait() {

	// keep waiting until the current wg value is nil
	for {
		var pw = p.gen.Load()
		if pw == nil {
			return // no mandated wait: noop return
		}
		pw.count.Add(1)
		pw.wg.Wait()
	}
}
