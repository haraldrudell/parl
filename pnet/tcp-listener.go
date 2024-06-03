/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"

	"github.com/haraldrudell/parl/perrors"
)

// TCPListener wraps [net.TCPListener]
//   - TCPListener provides:
//   - — thread-safe methods from [TCPListener] and [SocketListener]
//   - — promoted methods from [net.TCPListener]
//   - — all errors via unbound error channel
//   - — thread-safe state-handling for integrity
//   - — idempotent panic-free awaitable thread-safe [TCPListener.Close]
//   - — netip addresses
//   - — panic-handled handler connection threads receiving [*net.TCPConn]
//   - — awaitable Close and handler completion
type TCPListener struct {
	// the TCP IPv4 or IPv6 listening socket promoting TCP-specific methods
	//	- initialized by new-function, therefore thread-safe access
	net.TCPListener
	// connection-handler argument is *net.TCPConn
	SocketListener[*net.TCPConn]
}

// NewTCPListener returns an object for receiving tcp connections
//   - network is address family and type: NetworkTCP NetworkTCP4 NetworkTCP6, default IPv4
//   - socket implementation is [net.TCPListener]
//   - [SocketListener.Ch] returns a real-time unbound error channel or
//   - [SocketListener.Err] returns any errors appended into a single error
//   - [TCPListener.Close] is idempotent panic-free awaitable thread-safe
//   - [SocketListener.WaitCh]: channel closing on Close complete
//   - [SocketListener.Listen]: starts connection listening
//   - [SocketListener.AddrPort]: returns near socket address on successful listen
//   - [SocketListener.AcceptConnections]: provides connections to handler function until Close is invoked
//
// Usage:
//
//	var socket = pnet.NewTCPListener()
//	if err = r.socket.Listen("1.2.3.4:1234"); err != nil {
//		return // listen failed
//	}
//	defer parl.Close(&r.socket, &err)
//	println(socket.AddrPort().String()) // “127.0.0.1:1122”
func NewTCPListener(errp *error, network ...Network) (socket *TCPListener) {
	// single allocation here
	t := TCPListener{}
	var err error
	var addressType Network
	if len(network) > 0 {
		addressType = network[0]
	}
	if addressType == NetworkDefault {
		addressType = NetworkTCP4
	}
	NewSocketListener(&t.TCPListener, addressType, TransportTCP, &t.SocketListener, &err)
	if err != nil {
		*errp = perrors.AppendError(*errp, err)
		return
	}
	return &t
}

// - idempotent panic-free awaitable Close
func (s *TCPListener) Close() (err error) {
	return s.SocketListener.Close()
}
