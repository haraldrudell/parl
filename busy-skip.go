/*
© 2026–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "sync/atomic"

// BusySkip is a non-blocking first-win barrier
//
// Why:
//   - ensures a raised task to be eventually consistently exeucted exactly once
//     without ever blocking a thread
//   - similar to [sync.Mutex.TryLock] but more efficiently implemented by
//     a single, minimum-sized atomic
//
// Design:
//   - [BusySkip.EntryFailed] tries the barrier
//   - [BusySkip.Release] releases a held barrier
//   - 32-bit atomic, load-wrapped compare-and-swap
//   - written by Harald Rudell 260315
//
// Usage:
//
//	if busySkip.EntryFailed() {
//	  return
//	}
//	defer busySkip.Release()
type BusySkip struct {
	// state is [isFreeState] or [isBusyState]
	state atomic.Uint32
}

// IsBusy returns true if the busy-skip is busy. Thread-safe
func (b *BusySkip) IsBusy() (isBusy bool) { return b.state.Load() == isBusyState }

// IsFree returns true if the busy-skip is available. Thread-safe
func (b *BusySkip) IsFree() (isBusy bool) { return b.state.Load() == isFreeState }

// EntryFailed attempts to gain entry to the critical section
//   - wasBusy true: the invocation may not enter the critical section.
//     The ciritcal section was observed held by another thread
//   - wasBusy false: the ciritcal section was acquired for this thread
//   - — [BusySkip.Release] must be invoked
//   - —
//   - Thread-safe
//
// Usage:
//
//	if busySkip.EntryFailed() {
//	  return
//	}
//	defer busySkip.Release()
func (b *BusySkip) EntryFailed() (wasBusy bool) {
	return b.state.Load() == isBusyState || !b.state.CompareAndSwap(isFreeState, isBusyState)
}

// Release ends the critical section
//   - invoker must have received false from [BusySkip.EntryFailed]
//   - Thread-safe
func (b *BusySkip) Release() { b.state.Store(isFreeState) }

const (
	// the busy-skip is free
	isFreeState uint32 = iota
	// the busy-skip is busy
	isBusyState
)
