/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

// CloseChannel closes a channel handling panic, nil channel and
// optionally draining the channel prior to close
//   - ch: channel to close
//   - errp non-nil: receives any panic using [perrors.AppendError]
//   - drainChannel missing: no drain prior to close
//   - drainChannel [DoDrain]: drain channel prior to close.
//     If a thread is continuously sending items and DoDrain is present,
//     CloseChannel may block indefinitely.
//   - isNilChannel: true with no error when ch is nil.
//     Closing a nil channel is panic.
//   - isCloseOfClosedChannel: true with error if close paniced due to
//     the channel already closed.
//     A channel transferring data cannot be inspected for being
//     closed
//   - n: number of drained items with DoDrain
//   - Note: closing a channel while a thread is blocked in channel send is
//     a data race.
//   - thread-safe deferrable panic-free.
//     Handles closed-channel panic, nil-channel case and
//     has channel drain feature
func CloseChannel[T any](
	ch chan T,
	errp *error,
	drainChannel ...ChannelDrain,
) (
	isNilChannel, isCloseOfClosedChannel bool,
	n int,
	err error,
) {

	// nil channel case
	if isNilChannel = ch == nil; isNilChannel {
		return // closing of nil channel return
	}

	// channel drain feature
	if len(drainChannel) > 0 && drainChannel[0] {
		for {
			select {
			// read non-blocking from the channel
			//	- ok true: received item, channel is not closed
			//	- ok false: channel is closed
			case _, ok := <-ch:
				if ok {
					// the channel is not closed
					n++
					continue // read next item
				}
			default: // channel is open but has no items
			}
			break // closed or no items
		}
	}

	// close channel
	if Closer(ch, &err); err == nil {
		return // close successful return
	}

	// handle close error
	isCloseOfClosedChannel = pruntime.IsCloseOfClosedChannel(err)
	if errp != nil {
		*errp = perrors.AppendError(*errp, err)
	}

	return
}
