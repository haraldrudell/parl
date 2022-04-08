/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import "sync"

// ParlError is a thread-safe error container
type SendChannel struct {
	errCh        chan<- error // value and close inside lock
	shutdownOnce sync.Once
}

func NewSendChannel(errCh chan<- error) (sc *SendChannel) {
	return &SendChannel{errCh: errCh}
}

// Send sends an error on the error channel. Thread-safe
func (sc *SendChannel) Send(err error) {
	sc.errCh <- err // may block and panic
}

// Shutdown closes the channel exactly once. Thread-safe
func (sc *SendChannel) Shutdown() {
	sc.shutdownOnce.Do(func() {
		if sc.errCh == nil {
			return
		}
		close(sc.errCh)
	})
}
