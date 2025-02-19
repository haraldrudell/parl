/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"

	"github.com/haraldrudell/parl/perrors"
)

const (
	// minimum error channel length
	minGoResultLength = 1
	// default number of errors to receive
	defaultReceive = 1
)

// when GoResult.g is error channel only
type goResultChan chan error

func newGoResultChan(n ...int) (g goResultChan) {
	var n0 int
	if len(n) > 0 {
		n0 = n[0]
	}
	if n0 < minGoResultLength {
		n0 = minGoResultLength
	}
	return make(chan error, n0)
}

// SendError sends error as the final action of a goroutine
//   - SendError makes a goroutine:
//   - — awaitable and
//   - — able to return a fatal error
//   - — other needs of a goroutine is to initiate and detect cancel and
//     submit non-fatal errors
//   - errCh should be a buffered channel large enough for all its goroutines
//   - — this prevents goroutines from blocking in channel send
//   - SendError only panics from structural coding problems
//   - deferrable thread-safe
func (g goResultChan) SendError(errp *error) {
	if errp == nil {
		panic(NilError("errp"))
	}

	didSend, isNilChannel, isClosedChannel, err := ChannelSend(g, *errp, SendNonBlocking)
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

// ReceiveError is a deferrable function receiving error values from goroutines
//   - n is number of goroutines to wait for, default 1
//   - errp may be nil
//   - ReceiveError makes a goroutine:
//   - — awaitable and
//   - — able to return a fatal error
//   - — other needs of a goroutine is to initiate and detect cancel and
//     submit non-fatal errors
//   - GoRoutine should have enough capacity for all its goroutines
//   - — this prevents goroutines from blocking in channel send
//   - ReceiveError only panics from structural coding problems
//   - deferrable thread-safe
func (g goResultChan) ReceiveError(errp *error, n ...int) (err error) {
	var remainingErrors int
	if len(n) > 0 {
		remainingErrors = n[0]
	} else {
		remainingErrors = defaultReceive
	}

	// await goroutine results
	for ; remainingErrors > 0; remainingErrors-- {

		// blocks here
		//	- wait for a result from a goroutine
		var e = <-g
		if e == nil {
			continue // good return: ignore
		}

		// goroutine exited with error
		// ensure e has stack
		e = perrors.Stack(e)
		// build error list
		err = perrors.AppendError(err, e)
	}

	// final action: update errp if present
	if err != nil && errp != nil {
		*errp = perrors.AppendError(*errp, err)
	}

	return
}

// Count returns number of results that can be currently collected
//   - Thread-safe
func (g goResultChan) Count() (count int) { return len(g) }

func (g goResultChan) SetIsError() {
	panic(perrors.NewPF("NewGoResult does not provide SetIsError: use NewGoResult2"))
}

func (g goResultChan) IsError() (isError bool) {
	panic(perrors.NewPF("NewGoResult does not provide IsError: use NewGoResult2"))
}

func (g goResultChan) Remaining(add ...int) (adds, remaining int) {
	panic(perrors.NewPF("NewGoResult does not provide Remaining: use NewGoResult2"))
}

// “goResultChan0(1)”
func (g goResultChan) String() (s string) {
	return fmt.Sprintf("goResult_len:%d", len(g))
}
