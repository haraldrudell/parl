/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phttp

import (
	"net"
	"net/http"
	"sync"
	"sync/atomic"
)

// ListenerOnce wraps [net.Listener]
// making its Close method idempotent
type ListenerOnce struct {
	net.Listener
	isClose atomic.Bool
	wg      sync.WaitGroup
}

// ListenerOnce is [net.Listener]
var _ net.Listener = &ListenerOnce{}

// MakeListenerOnce returns [net.Listener] with
// idempotent Close method
//   - —
//   - this is sometimes required with [http.Server]
//   - idempotent Close means a deferred Close can be invoked
//     when an error state has made it uncertain whether Close
//     was already invoked
//   - [http.Server.Serve] has package-private implementation
func MakeListenerOnce(listener net.Listener) (o ListenerOnce) {

	o = ListenerOnce{
		Listener: listener,
	}
	o.wg.Add(1)

	return
}

// Close: only the first Close is effective and may receive error
//   - no invocation returns prior to the listener being closed
//   - error is never returned for subsequent threads
func (o *ListenerOnce) Close() (err error) {

	// pick winner
	if o.isClose.Load() || !o.isClose.CompareAndSwap(false, true) {

		// losers wait
		o.wg.Wait()
		return
	}
	defer o.wg.Done()

	// winner closes
	if o.Listener == nil {
		return
	}
	err = o.Listener.Close()

	return
}

var _ = (&http.Server{}).Serve
