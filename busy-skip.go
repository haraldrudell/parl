/*
© 2026–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "sync/atomic"

// BusySkip is a wait-free critical section
//
// Usage:
//
//	if !q.stateBusySkip.Seek() {
//	  return
//	}
//	defer q.stateBusySkip.Release()
type BusySkip struct {
	state atomic.Uint32
}

func (b *BusySkip) IsBusy() (isBusy bool) { return b.state.Load() == isBusyState }

func (b *BusySkip) IsFree() (isBusy bool) { return b.state.Load() == isFreeState }

// EntryFailed attempts to gain entry to the critical section
//   - granted: the ciritcal section was availale
//
// Usage:
//
//	if !q.stateBusySkip.EntryFailed() {
//	  return
//	}
//	defer q.stateBusySkip.Release()
func (b *BusySkip) EntryFailed() (rejected bool) {
	return b.state.Load() == isBusyState || !b.state.CompareAndSwap(isFreeState, isBusyState)
}

// Release ends the critical section
func (b *BusySkip) Release() { b.state.Store(isFreeState) }

const (
	// the busy-skip is free
	isFreeState uint32 = iota
	// the busy-skip is busy
	isBusyState
)
