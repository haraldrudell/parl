/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"sync"
)

// ParlError is a thread-safe error container
type SendChannel struct {
	errChLock sync.Mutex
	errCh     chan<- error // value and close inside lock
	panicFunc func(err error)
}

func NewSendChannel(errCh chan<- error, panicFunc func(err error)) (sc *SendChannel) {
	if panicFunc == nil {
		panicFunc = defaultPanic
	}
	return &SendChannel{errCh: errCh, panicFunc: panicFunc}
}

// Send sends an error on the error channel. Thread-safe
func (sc *SendChannel) Send(err error) {
	sc.getErrCh() <- err // may block and panic
}

// Shutdown closes the channel exactly once. Thread-safe
func (sc *SendChannel) Shutdown() {
	defer RecoverThread("ParlError panic on closing errCh", sc.panicFunc)

	sc.errChLock.Lock()
	defer sc.errChLock.Unlock()
	errCh := sc.errCh
	if errCh == nil {
		return
	}
	sc.errCh = nil // prevent further send or close
	close(errCh)
}

func (sc *SendChannel) getErrCh() (errCh chan<- error) {
	sc.errChLock.Lock()
	defer sc.errChLock.Unlock()
	return sc.errCh
}

func defaultPanic(err error) {
	panic(err)
}
