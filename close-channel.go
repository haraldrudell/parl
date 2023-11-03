/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	CloseChannelDrain = true
)

// CloseChannel closes a channel
//   - CloseChannel is thread-safe, deferrable and panic-free,
//     handles closed-channel panic, nil-channel case and
//     has channel drain feature
//   - isNilChannel returns true if ch is nil.
//     closing a nil channel would cause panic.
//   - isCloseOfClosedChannel is true if close paniced due to
//     the channel already closed.
//     A channel transferring data cannot be inspected for being
//     closed
//   - if errp is non-nil, panic values updates it using errors.AppendError.
//   - if doDrain is [parl.CloseChannelDrain], the channel is drained first.
//     Note: closing a channel while a thread is blocked in channel send is
//     a data race.
//     If a thread is continuously sending items and doDrain is true,
//     CloseChannel will block indefinitely.
//   - n returns the number of drained items.
func CloseChannel[T any](ch chan T, errp *error, drainChannel ...bool) (
	isNilChannel, isCloseOfClosedChannel bool, n int, err error,
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
