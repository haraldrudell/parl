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

// Listen obtains a tcp or other listener
//   - network must be set.
//     NetworkDefault yields “panic: listen: unknown network”
//   - addr can be empty
//   - default port is ephemeral, a high number like 59321
//   - if IPv6 is allowed, default host typically becomes “::” not “::1”
//   - limited to IPv4, eg. NetworkTCP4, default host typically becomes “0.0.0.0” not “127.0.0.1”
func Listen(
	network Network, addr string,
	cancel *atomic.Pointer[context.CancelFunc],
) (listener net.Listener, err error) {
	var ctx, cancelFunc = context.WithCancel(context.Background())
	defer cancelFunc()
	cancel.Store(&cancelFunc)
	defer cancel.Store(nil)

	var listenConfig = net.ListenConfig{}
	listener, err = listenConfig.Listen(ctx, network.String(), addr)
	if err != nil {
		err = perrors.ErrorfPF("net.Listen %s %s: '%w'", network, addr, err)
	}

	return
}
