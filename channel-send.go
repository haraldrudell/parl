/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

const SendNonBlocking = true // with nonBlocking set to SendNonBlocking, ChannelSend will never block

// ChannelSend is channel send without panics and possibly non-blocking
//   - if nonBlocking is SendNonBlocking or true, channel send will be attempted but not block
//   - didSend is true if value was successfully sent on ch
//   - err is non-nil if a panic occurred or ch is nil
//   - isNilChannel is true if the channel is nil, ie. send would block indefinitely
//   - isClosedChannel is true if the panic was caused by ch being closed
//   - there should be no panics other than from ch being closed
func ChannelSend[T any](ch chan<- T, value T, nonBlocking ...bool) (didSend, isNilChannel, isClosedChannel bool, err error) {
	var sendNb bool
	if len(nonBlocking) > 0 {
		sendNb = nonBlocking[0]
	}
	defer Recover(Annotation(), &err, func(e error) {
		if pruntime.IsSendOnClosedChannel(e) {
			isClosedChannel = true
		}
	})

	if isNilChannel = ch == nil; isNilChannel {
		err = perrors.NewPF("ch channel nil")
		return
	}

	// send non-blocking
	if sendNb {
		select {
		case ch <- value:
			didSend = true
		default:
		}
		return
	}

	// send blocking: blocks here
	ch <- value
	didSend = true

	return
}
