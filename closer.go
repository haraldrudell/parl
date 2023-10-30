/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"io"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	CloseChannelDrain = true
)

// Closer is a deferrable function that closes a channel.
// Closer handles panics.
// if errp is non-nil, panic values updates it using errors.AppendError.
func Closer[T any](ch chan T, errp *error) {
	defer PanicToErr(errp)

	close(ch)
}

// CloserSend is a deferrable function that closes a send-channel.
// CloserSend handles panics.
// if errp is non-nil, panic values updates it using errors.AppendError.
func CloserSend[T any](ch chan<- T, errp *error) {
	defer PanicToErr(errp)

	close(ch)
}

// Close is a deferrable function that closes an io.Closer object.
// Close handles panics.
// if errp is non-nil, panic values updates it using errors.AppendError.
func Close(closable io.Closer, errp *error) {
	defer PanicToErr(errp)

	if e := closable.Close(); e != nil {
		*errp = perrors.AppendError(*errp, e)
	}
}

// CloseChannel closes a channel recovering panics
//   - deferrable
//   - if errp is non-nil, panic values updates it using errors.AppendError.
//   - if doDrain is CloseChannelDrain or true, the channel is drained first.
//     Note: closing a channel while a thread is blocked in channel send is
//     a data race.
//     If a thread is continuously sending items and doDrain is true,
//     CloseChannel will block indefinitely.
//   - n returns the number of drained items.
//   - isNilChannel returns true if ch is nil.
//     No close will be attempted for a nil channel, it would panic.
func CloseChannel[T any](ch chan T, errp *error, drainChannel ...bool) (
	isNilChannel, isCloseOfClosedChannel bool, n int, err error,
) {
	if isNilChannel = ch == nil; isNilChannel {
		return // closing of nil channel return
	}
	var doDrain bool
	if len(drainChannel) > 0 {
		doDrain = drainChannel[0]
	}
	if doDrain {
		var hasItems = true
		for hasItems {
			select {
			// read non-blocking from the channel
			case _, ok := <-ch:
				if ok {
					// the channel is not closed
					n++
					continue // read next item
				}
			default:
			}
			hasItems = false
		}
	}
	Closer(ch, &err)
	if err == nil {
		return // close successful
	}
	isCloseOfClosedChannel = pruntime.IsCloseOfClosedChannel(err)
	if errp != nil {
		*errp = perrors.AppendError(*errp, err)
	}
	return
}
