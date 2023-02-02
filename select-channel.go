/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
)

// SelectChannel provides a channel that closes upon the next Trig. Thread-safe.
//   - SelectChannel is designed to be used in an external select statement receiving from
//     contexts, the select-channel and other data-providing channels
//   - SelectChannel allows a thread waiting in external sleect-statement to be:
//   - — cancelled via contexts,
//   - — receive data or
//   - — be notified of alternative action via the select-channel
//   - SelectChannel allows many threads to wait until an action is taken by another thread
//   - SelectChannel does not require initialization
type SelectChannel struct {
	chLock sync.Mutex
	ch     chan struct{}
}

// NewSelectChannel returns a thread-safe provider of channels that close upon Trig
func NewSelectChannel() (channelWaiter *SelectChannel) {
	return &SelectChannel{}
}

// Ch provides a context-like channel for a select and this provided channel closes on Trig. Thread-safe
func (cw *SelectChannel) Ch() (ch <-chan struct{}) {
	cw.chLock.Lock()
	defer cw.chLock.Unlock()

	if ch = cw.ch; ch == nil {
		cw.ch = make(chan struct{})
		ch = cw.ch
	}
	return
}

// Trig closes the current channel and initializes a new channel. Thread-safe
func (cw *SelectChannel) Trig() {
	cw.chLock.Lock()
	defer cw.chLock.Unlock()

	oldCh := cw.ch
	if oldCh == nil {
		return
	}

	close(oldCh)
	cw.ch = nil
}
