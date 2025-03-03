/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import "sync/atomic"

// OnceCh implements a one-time execution filter
//   - initialization free
//   - OnceCh is similar to [sync.Once] with improvements:
//   - — does not require an awkward function value to be provided.
//     Method dunction-values cause allocation
//   - — awaitable channel mechanic means threads can
//     await multiple events
//   - — is observable via [OnceCh.IsInvoked] [OnceCh.IsClosed]
//
// Usage:
//
//	var o OnceCh
//	if isWinner, done := o.IsWinner(); !isWinner {
//	  return // thread already waited for winner thread completion
//	} else {
//	  defer done.Done()
//	}
//	…
type OnceCh struct {
	// winner selects the winning thread
	winner atomic.Bool
	// done allows:
	//	- winner thread to indicate completion
	//	- loser threads to await winner completion
	//	- mechanic: closing channel
	//	- separate struct provided to winner as Done implementation
	done doneOnce
}

// doneOnce provides single exported method Done to winner only
type doneOnce struct {
	// awaitable is mechanic for loser threads to
	// await winner execution complete
	awaitable Awaitable
}

// doneOnce implements Done interface
var _ Done = &doneOnce{}

// Done allows the winner to signal completion
func (d *doneOnce) Done() { d.awaitable.Close() }

// IsWinner selects winner thread as the first of invokers
//   - noWait missing: loser thread wait for winner thread invoking [done.Done]
//   - noWait NoOnceWait: eventually consistent: loser threads immediately return
//   - isWinner true: invoker thread is winner, ie. the first invoker of [OnceCh.IsWinner]
//   - — must invoke [done.Done] upon task completion
//   - isWinner false: loser thread, done is nil
//   - done: non-nil for winner only. Provides [done.Done] method
func (o *OnceCh) IsWinner(noWait ...OnceChStrategy) (isWinner bool, done Done) {

	// pick winner thread
	if isWinner = !o.winner.Load() && o.winner.CompareAndSwap(false, true); isWinner {
		done = &o.done
		return // winner return
	}

	// loser threads wait for winner thread unless noWait: NoOnceWait
	if len(noWait) == 0 || !noWait[0] {
		<-o.done.awaitable.Ch()
	}

	return
}

// Ch returns a channel that closes once IsWinner and done have both been invoked
func (o *OnceCh) Ch() (ch AwaitableCh) { return o.done.awaitable.Ch() }

// IsInvoked indicates that a winner was selected
func (o *OnceCh) IsInvoked() (isInvoked bool) { return o.winner.Load() }

// IsClosed indicates that a winner was selected and the winneer invoked done
func (o *OnceCh) IsClosed() (isClosed bool) { return o.done.awaitable.IsClosed() }
