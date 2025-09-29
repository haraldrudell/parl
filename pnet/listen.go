/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"context"
	"net"
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
)

// Listen obtains a tcp or stream-oriented domain-socket listener
//   - for connection-oriented and their derived protocols like https
//   - socketAddress: contains:
//   - — [Network] tcp/tcp4/tcp6 and either
//   - — an IPv4/IPv6 [netip.AddrPort] literal or
//   - — domain name resolving to a local interface: “example.com:1234”
//   - — domain-socket address “/socket” “@socket”
//   - — for tcp, zero port number selects an ephemeral port
//   - — if IPv6 is supported, “localhost” typically becomes “::” not “::1”
//   - cancel: optional pointer that is set to a cancel function
//     during listen invocation
//   - err: socketAddr is invalid domain string like ‘%s’ or netip.Addr nil value
//   - —
//   - delegates to [net.ListenConfig.Listen] that supports context cancel
//   - network value is not used by the kernel, it is a standard-library scoped
//     helper
//   - TODO 240616 possibly refactor cancel argument
func Listen(
	socketAddress SocketAddress,
	cancel ...*atomic.Pointer[context.CancelFunc],
) (listener net.Listener, err error) {

	// [net.ListenConfig.Listen] requires a context
	var ctx = context.Background()

	// handle cancel argument
	if len(cancel) > 0 {
		if cancel0 := cancel[0]; cancel0 != nil {
			var cancelFunc context.CancelFunc
			ctx, cancelFunc = context.WithCancel(ctx)
			defer cancelFunc()
			cancel0.Store(&cancelFunc)
			defer cancel0.Store(nil)
		}
	}

	var listenConfig = net.ListenConfig{}
	var network = socketAddress.Network().String()
	var addr = socketAddress.String()
	// listen to empty string becomes TCPAddr with IP length 0 port 0
	// which becomes any interface with high ephemeral port
	if listener, err = listenConfig.Listen(ctx, network, addr); err != nil {
		err = perrors.ErrorfPF("net.Listen %s %s: “%w”", network, addr, err)
	}

	return
}
