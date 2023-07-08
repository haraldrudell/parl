/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/perrors"

// ErrChWait is a deferrable function receiving an error value on a channel
//   - used to wait for a goroutine
//
// Usage:
//
//	var err error
//	defer errorHandler(&err)
//
//	var errCh = make(chan error, 1)
//	go someFunc(errCh)
//	defer parl.ErrChWait(errCh, &err)
//
//	func someFunc(errCh chan<- error) {
//	  var err error
//	  defer parl.SendErr(errCh, &err)
//	  defer parl.Recover(parl.Annotation(), &err, parl.NoOnError)
func ErrChWait(errCh <-chan error, errp *error) {
	if errp == nil {
		panic(perrors.NewPF("errp cannot be nil"))
	} else if errCh == nil {
		panic(perrors.NewPF("errCh cannot be nil"))
	}
	// blocks here
	if err := <-errCh; err != nil {
		*errp = perrors.AppendError(*errp, err)
	}
}

// SendErr sends error as the final action of a goroutine
//   - SendErr should only panic from structural coding problems
func SendErr(errCh chan<- error, errp *error) {
	if errp == nil {
		panic(perrors.NewPF("errp cannot be nil"))
	}
	didSend, isNilChannel, isClosedChannel, err := ChannelSend(errCh, *errp, SendNonBlocking)
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
