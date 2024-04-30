/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import "github.com/haraldrudell/parl/sets"

// SocketTransport describes what type of socket from the types supported by Go
//   - TransportTCP TransportUDP TransportIP TransportUnix
//   - 0 is uninitalized invalid
type SocketTransport uint8

const (
	// tcp connection oriented
	//	- listener implementation is *net.TCPListener
	//	- network is tcp tcp4 tcp6
	//	- netListener.Addr implementation is *net.TCPAddr
	TransportTCP = iota + 1
	// udp connectionless
	TransportUDP
	// ip: no port: protocols like icmp
	TransportIP
	// Unix socket inside the kernel to a process on the same host
	TransportUnix
)

func (t SocketTransport) String() (s string) {
	return transportSet.StringT(t)
}

// IsValid returns true if the SocketTransport has been initialized
func (t SocketTransport) IsValid() (isValid bool) {
	return transportSet.IsValid(t)
}

// transportSet is the set variable for SocketTransport
var transportSet = sets.NewSet[SocketTransport]([]sets.SetElement[SocketTransport]{
	{ValueV: TransportTCP, Name: "tcp"},
})
