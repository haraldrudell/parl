/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"

	"github.com/haraldrudell/parl/perrors"
)

// goResultChan is minimum [GoResult] implementation: a channel with methods
type goResultChan chan error

// newGoResultChan returns minimum [GoResult] implementation
func newGoResultChan(n ...int) (g goResultChan) {

	// get n
	var n0 int
	if len(n) > 0 {
		n0 = n[0]
	}
	if n0 < minGoResultLength {
		n0 = minGoResultLength
	}

	g = make(chan error, n0)

	return
}

// Done sends error as the final action of a goroutine
//   - SendError makes a goroutine:
//   - — awaitable and
//   - — able to return a fatal error
//   - — other needs of a goroutine is to initiate and detect cancel and
//     submit non-fatal errors
//   - errCh should be a buffered channel large enough for all its goroutines
//   - — this prevents goroutines from blocking in channel send
//   - SendError only panics from structural coding problems
//   - deferrable thread-safe
func (g goResultChan) done(err0 error) {
	// send should never fail. If it does: panic
	var didSend, isNilChannel, isClosedChannel, err = ChannelSend(g, err0, SendNonBlocking)
	if didSend {
		return // error value sent return
	} else if isNilChannel {
		err = perrors.ErrorfPF("fatal: error channel nil: %w", err)
	} else if isClosedChannel {
		err = perrors.ErrorfPF("fatal: error channel closed: %w", err)
	} else if err != nil {
		err = perrors.ErrorfPF("fatal: panic when sending on error channel: %w", err)
	} else {
		err = perrors.NewPF("fatal: error channel blocking on send")
	}
	panic(err)
}

// ch obtains the error providing channel
func (g goResultChan) ch() (ch <-chan error) { return g }

// count returns number of results that can be currently collected
//   - Thread-safe
func (g goResultChan) count() (available, stillRunning int) {
	available = len(g)
	return
}

// “goResultChan0(1)”
func (g goResultChan) String() (s string) {
	return fmt.Sprintf("goResult_len:%d", len(g))
}

const (
	// minimum error channel length
	minGoResultLength = 1
)
