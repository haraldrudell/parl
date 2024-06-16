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

var NoCancel *atomic.Pointer[context.CancelFunc]

// Listen obtains a tcp or other listener
//   - socketAddress: contains:
//   - — [Network] and either
//   - — an IPv4/IPv6 [netip.AddrPort] literal or
//   - — domain name resolving to a local interface: “example.com:1234”
//   - — domain-socket address “/socket” “@socket”
//   - — zero port number selects an ephemeral port
//   - — if IPv6 is allowed, default host typically becomes “::” not “::1”
//   - — cancel is an optional pointer that is set to a cancel function
//     during listen invocation.
//     May have value [NoCancel]
func Listen(
	socketAddress SocketAddress,
	cancel *atomic.Pointer[context.CancelFunc],
) (listener net.Listener, err error) {
	var ctx, cancelFunc = context.WithCancel(context.Background())
	defer cancelFunc()
	if cancel != nil {
		cancel.Store(&cancelFunc)
		defer cancel.Store(nil)
	}

	var listenConfig = net.ListenConfig{}
	var network = socketAddress.Network().String()
	var addr = socketAddress.String()
	listener, err = listenConfig.Listen(ctx, network, addr)
	if err != nil {
		err = perrors.ErrorfPF("net.Listen %s %s: '%w'", network, addr, err)
	}

	return
}
