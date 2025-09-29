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

	"github.com/haraldrudell/parl"
)

// ListenerOnce wraps [net.Listener]
// making its Close method idempotent
type ListenerOnce struct {
	// net.Listener provides promoted method Addr()
	net.Listener
	// isClose is true if Close has been invoked
	//	- atomic performance
	isClose atomic.Bool
	// wg.Wait awaits close completion
	//	- for Close loser threads
	wg sync.WaitGroup
	// limiter limits pending Accept threads across
	// multiple listeners
	//	- nil: no limit
	limiter parl.Moderate
}

// ListenerOnce is [net.Listener]
var _ net.Listener = &ListenerOnce{}

// MakeListenerOnce returns [net.Listener] with
// idempotent Close method and optional limiting of parallelism
//   - —
//   - [http.Server] design of Close Shutdown closes listeners multiple times
//   - if Close were not idempotent, multiple Close is panic
//   - with ListenerOnce, a deferred Close can be invoked regardless of state
//   - [http.Server.Serve] has package-private implementation
func MakeListenerOnce(listener net.Listener, limiter ...parl.Moderate) (o ListenerOnce) {

	o = ListenerOnce{
		Listener: listener,
	}
	if len(limiter) > 0 {
		o.limiter = limiter[0]
	}
	// One means Close did not complete yet
	o.wg.Add(1)

	return
}

// Accept delegates to [net.Listener.Accept] possibly blocked
// by awaiting limiter ticket
func (o *ListenerOnce) Accept() (netConn net.Conn, err error) {
	if o.limiter != nil {
		defer o.limiter.Ticket().ReturnTicket()
	}
	return o.Listener.Accept()
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
