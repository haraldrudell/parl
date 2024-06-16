/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"fmt"
	"net/netip"
)

// AddrPort may hold address and port for IP, TCP, UDP, domain-socket
// or derived protocols
//   - used for listeners or outbound connections
//   - for TCP UDP IP networks:
//   - — address empty or
//   - — literal unspecified IP address:
//   - — Listen listens on all available unicast and anycast IP addresses
//   - port 0: listener picks an ephemeral port
//   - implemented by [netip.AddrPort]
//   - created by [NewAddrPort]
//   - also:
//   - — [Address]
//   - — [SocketAddress]
//   - — [Network]
type AddrPort interface {
	// socket address as [netip.AddrPort]
	//	- if no IP literal exists, [netip.AddrPort.IsValid] is false
	AddrPort() (addrPort netip.AddrPort)
	// string representation of this address. May be:
	//	- string value of [SocketAddress.AddrPort]: “1.2.3.4:5” “[::1]:2”
	//	- a domain name “example.com:1234”
	//	- a domain-socket address “/socket” “@socket”
	//	- empty string if domain is empty or AddrPort is invalid
	fmt.Stringer
}

func NewAddrPort(domainPort string) (addrP AddrPort) { return &addrPort{domainPort: domainPort} }

type addrPort struct{ domainPort string }

func (a addrPort) AddrPort() (addrPort netip.AddrPort) { return }

func (a addrPort) String() (s string) { return a.domainPort }
