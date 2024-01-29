/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import (
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
)

const (
	bTrigBit     = 0x1
	bNegativeBit = 0x2
	bShift       = 2
)

type CountingAwaitable struct {
	Awaitable
	counter atomic.Int64
	adds    atomic.Uint64
}

// NewCountingAwaitable returns a counting awaitable
//   - similar to sync.WaitGroup but wait-mechanic is closing channel
//   - observable via Count method.
//     Count with missing delta will not panic
//   - —
//   - state is determined by counter.
//     Guaranteed state is returned by Count IsTriggered
//   - IsClosed is eventually consistent.
//     A parallel Count may reflect triggered prior IsClose returning true.
//     Upon return from the effective Count Add Done IsTriggered invocation,
//     IsClose is consistent and the channel is closed if it is to close.
//     For race-sensitive code, synchronize or rely on Count
//   - similarly, a parallel Count invocation may reflect triggered state
//     prior to the Awaitable channel actually closing
//   - mechanic is atomic.Int64.CompareAndSwap
func NewCountingAwaitable() (awaitable *CountingAwaitable) {
	return &CountingAwaitable{}
}

// IsTriggered returns true if counter has been used and returned to zero
//   - IsClosed true, AwaitableCh closed
func (a *CountingAwaitable) IsTriggered() (isTriggered bool) {
	return a.counter.Load()&bTrigBit != 0
}

// Count returns the current count and may also adjust the counter
//   - counter is current remaining count
//   - adds is cumulative positive adds
//   - if delta is present and the awaitable is triggered, this is panic
//   - state can be retrieved without panic by omitting delta
//   - if delta is negative and result is zero: trig
//   - if delta is negative and result is negative: panic
//   - similar to sync.WaitGroup.Add Done
func (a *CountingAwaitable) Count(delta ...int) (counter, adds int) {
	var hasDelta = len(delta) > 0
	if !hasDelta {
		counter = int(a.counter.Load() >> bShift)
		adds = int(a.adds.Load())
		return
	}
	var delta0 = delta[0]

	// thread-safe add to counter
	var counter64 int64
	for {
		var value = a.counter.Load()
		if value&bNegativeBit != 0 {
			panic(perrors.NewPF("negative fail-state encountered"))
		} else if value&bTrigBit != 0 {
			panic(perrors.NewPF("Add or Done on triggered CountingAwaitable"))
		}
		if delta0 == 0 {
			counter = int(value >> bShift)
			adds = int(a.adds.Load())
			return // noop
		}
		counter64 = value + int64(delta0)<<bShift
		if counter64 == 0 {
			counter64 |= bTrigBit
		} else if counter < 0 {
			counter |= bNegativeBit
		}
		if a.counter.CompareAndSwap(value, counter64) {
			break // add succeeded
		}
	}
	counter = int(counter64 >> bShift)
	if delta0 > 0 {
		adds = int(a.adds.Add(uint64(delta0)))
	} else {
		adds = int(a.adds.Load())
	}
	if counter64 != bTrigBit && counter > 0 {
		return // no trig or negative
	} else if counter64 < 0 {
		panic(perrors.NewPF("negative counter"))
	}

	// trig!
	if !a.Close() {
		panic(perrors.NewPF("Awaitable already triggered"))
	}

	return
}

// Add signals thread-launch by adding to the counter
//   - if delta is negative and result is zero: trig
//   - if delta is negative and result is negative: panic
//   - Add on triggered or negative is panic
//   - similar to sync.WaitGroup.Add
func (a *CountingAwaitable) Add(delta int) { a.Count(delta) }

// Done signals thread exit by decrementing the counter
//   - if delta is negative and result is zero: trig
//   - if delta is negative and result is negative: panic
//   - Add on triggered or negative is panic
//   - similar to sync.WaitGroup.Done
func (a *CountingAwaitable) Done() { a.Count(-1) }
