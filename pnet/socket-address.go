/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"fmt"
	"net/netip"
)

// SocketAddress may hold address and port for TCP, UDP
// or derived protocols
// - a socket address consists of:
//   - — transport protocol
//   - — IP address or domain
//   - — port number
//   - used for listeners or outbound connections
//   - for TCP UDP networks:
//   - — address empty or
//   - — literal unspecified IP address:
//   - — Listen listens on all available unicast and anycast IP addresses
//   - port 0: listener picks an ephemeral port
//   - created by [NewSocketAddressLiteral] [NewSocketAddress]
//   - also:
//   - — [Address]
//   - — [AddrPort]
//   - — [Network]
type SocketAddress interface {
	// network option, only returned in this method
	Network() (network Network)
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

func NewSocketAddressFromString(network Network, addr string) (socketAddress SocketAddress) {
	if a, e := netip.ParseAddrPort(addr); e == nil {
		socketAddress = NewSocketAddressLiteral(network, a)
		return
	}
	socketAddress = NewSocketAddress(network, addr)

	return
}

func NewSocketAddressLiteral(network Network, addrPort netip.AddrPort) (socketAddress SocketAddress) {
	return &saLit{network: network, addrPort: addrPort}
}

func NewSocketAddress(network Network, domain string) (socketAddress SocketAddress) {
	return &sa{network: network, domain: domain}
}

type saLit struct {
	network  Network
	addrPort netip.AddrPort
}

func (a saLit) AddrPort() (addrPort netip.AddrPort) { return a.addrPort }
func (a saLit) Network() (network Network)          { return a.network }
func (a saLit) String() (s string)                  { return a.addrPort.String() }

type sa struct {
	network Network
	domain  string
}

func (a sa) AddrPort() (addrPort netip.AddrPort) { return }
func (a sa) Network() (network Network)          { return a.network }
func (a sa) String() (s string)                  { return a.domain }
