/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import "sync"

// ParlError is a thread-safe error container
type SendChannel struct {
	errCh        chan<- error // value and close inside lock
	onError      func(err error)
	shutdownOnce sync.Once
}

func NewSendChannel(errCh chan<- error, onError func(err error)) (sc *SendChannel) {
	if onError == nil {
		onError = defaultPanic
	}
	return &SendChannel{errCh: errCh, onError: onError}
}

// Send sends an error on the error channel. Thread-safe
func (sc *SendChannel) Send(err error) {
	sc.errCh <- err // may block and panic
}

// Shutdown closes the channel exactly once. Thread-safe
func (sc *SendChannel) Shutdown() {
	sc.shutdownOnce.Do(func() {
		defer RecoverThread("ParlError panic on closing errCh", sc.getPanicFunc())

		if sc.errCh == nil {
			return
		}
		close(sc.errCh)
	})
}

func (sc *SendChannel) getPanicFunc() func(err error) {
	if sc.onError != nil {
		return sc.onError
	}
	return defaultPanic
}

func defaultPanic(err error) {
	panic(err)
}
