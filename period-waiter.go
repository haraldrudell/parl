/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"math"
	"sync"
	"sync/atomic"
)

// PeriodWaiter blocks Wait invokers while a HoldWaiters
// that has not been succeeded by a ReleaseWaiters invocation. Thread-safe
type PeriodWaiter struct {
	// wg holds Wait invokers
	//	- wg is atomically updated
	wg    atomic.Pointer[sync.WaitGroup]
	count uint64 // atomic. number of threads in Wait
}

// NewPeriodWaiter returns an object that can hold threads periodically. Thread-safe
func NewPeriodWaiter() (periodWaiter *PeriodWaiter) {
	return &PeriodWaiter{}
}

// HoldWaiters causes a thread invoking [PeriodWaiter.Wait] to wait. Thread-safe
func (p *PeriodWaiter) HoldWaiters() {
	if p.wg.Load() != nil {
		return // already waiting return: noop
	}
	var wg sync.WaitGroup
	wg.Add(1)
	// atomic nil to value of p.wg
	p.wg.CompareAndSwap(nil, &wg)
}

// ReleaseWaiters releases any threads blocked in [PeriodWaiter.Wait]
// and lets new Wait invokers proceed. Thread-safe
func (p *PeriodWaiter) ReleaseWaiters() {
	// atomic read of the current wg
	var wg = p.wg.Load()
	if wg == nil {
		return // state not waiting return: noop
	}

	// resolve any pending waitgroup
	if !p.wg.CompareAndSwap(wg, nil) {
		return // other thread already niled it: noop
	}
	wg.Done()
}

// Count returns the number of threads currently in Wait
func (p *PeriodWaiter) Count() (waitingThreads int) {
	return int(atomic.LoadUint64(&p.count))
}

// IsHold returns whether Wait will currently block
func (p *PeriodWaiter) IsHold() (isHold bool) {
	return p.wg.Load() != nil
}

// Wait blocks the thread if a HoldWaiters invocation took place with no
// ReleaseWaiters succeeding it. Thread-safe
func (p *PeriodWaiter) Wait() {
	// first atomic read of current p.wg
	var wg = p.wg.Load()
	if wg == nil {
		return // no mandated wait return
	}

	atomic.AddUint64(&p.count, 1)
	defer atomic.AddUint64(&p.count, math.MaxUint64)

	// keep waiting until the current wg value is nil
	for {
		wg.Wait()
		if wg = p.wg.Load(); wg == nil {
			return // wait complete return
		}
	}
}
