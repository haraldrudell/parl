/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import "sync/atomic"

const (
	// [OnceCh.IsWinner]
	NoOnceWait = true
)

// OnceCh implements a one-time execution filter
//   - initialization free
//
// Usage:
//
//	var o OnceCh
//	if isWinner, closeFunc := o.IsWinner(); !isWinner {
//	  return // thread already waited for winner thread completion
//	} else {
//	  defer closeFunc()
//	}
//	…
type OnceCh struct {
	winner    atomic.Bool
	awaitable Awaitable
}

// IsWinner returns true for the first invoker
//   - subsequent invokers wait for the awaitable then return false
//   - if noWait is NoOnceWait, loser threads do not wait
//   - isWinner true has closeFunc non-nil
func (o *OnceCh) IsWinner(noWait ...bool) (isWinner bool, closeFunc func()) {

	// pick winner thread
	if isWinner = o.winner.CompareAndSwap(false, true); isWinner {
		closeFunc = o.close
		return // winner return
	}

	// loser threads wait for winner thread unless noWait: NoOnceWait
	if len(noWait) == 0 || !noWait[0] {
		<-o.awaitable.Ch()
	}

	return
}

func (o *OnceCh) Ch() (ch AwaitableCh) { return o.awaitable.Ch() }

func (o *OnceCh) IsInvoked() (isInvoked bool) { return o.winner.Load() }

func (o *OnceCh) IsClosed() (isClosed bool) { return o.awaitable.IsClosed() }

func (o *OnceCh) close() { o.awaitable.Close() }
