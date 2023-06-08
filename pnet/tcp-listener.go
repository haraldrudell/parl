/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"
)

// TCPListener embeds net.TCPListener
//   - panic-handled connection threads receiving *net.TCPConn
//   - Ch: real-time error channel or
//   - Err: collecting errors after close
//   - idempotent panic-free awaitable Close
//   - WaitCh: channel closing on Close complete
type TCPListener struct {
	*net.TCPListener             // the TCP IPv4 or IPv6 listening socket promoting TCP-specific methods
	SocketListener[*net.TCPConn] // connection-handler argument is *net.TCPConn
}

// NewTCPListener returns object for receiving IPv4 tcp connections
func NewTCPListener(network ...Network) (socket *TCPListener) {
	var defaultNetwork Network
	if len(network) > 0 {
		defaultNetwork = network[0]
	} else {
		defaultNetwork = NetworkTCP4
	}

	var listener net.TCPListener
	return &TCPListener{
		TCPListener:    &listener,
		SocketListener: *NewSocketListener[*net.TCPConn](&listener, defaultNetwork, TransportTCP),
	}
}

// - idempotent panic-free awaitable Close
func (s *TCPListener) Close() (err error) {
	return s.SocketListener.Close()
}
