/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/haraldrudell/parl/perrors"
)

// SocketAddress may hold address and port for TCP, UDP
// or other protocols. Used to generically hold a near or far endpoint
// for net listeners and clients. Uses go1.18 2022 [netip.AddrPort].
//   - a socket address can be defined by IP literal or text domain-name
//   - created by:
//   - — [NewSocketAddressFromString] “1.2.3.4:5” or “example.com:3”
//   - — [NewSocketAddressLiteral] “1.2.3.4:5”
//   - — [NewSocketAddress] “example.com:3”
//   - a socket address consists of:
//   - — transport protocol
//   - — IP address or domain
//   - — port number
//   - for TCP UDP networks:
//   - — address empty or
//   - — literal unspecified IP address:
//   - — Listen listens on all available unicast and anycast IP addresses
//   - — port 0: listener picks an ephemeral port
//     49152–65535, Linux 32768–60999
//   - related types:
//   - — [Address] is similar but for address only
//     rather than address and port
//   - — [AddrPort] holds domain-based socket address
//     with no IP literal or network protocol
//   - — [Network] type-safe network protocol definitions
type SocketAddress interface {
	// Network returns the network protocol “tcp” “udp6”
	//	- only exposed by this method
	Network() (network Network)
	// AddrPort returns socket address as [netip.AddrPort]
	//	- if no IP literal exists, [netip.AddrPort.IsValid] is false
	AddrPort() (addrPort netip.AddrPort)
	// string representation of this address. May be:
	//	- string value of [SocketAddress.AddrPort]: “1.2.3.4:5” “[::1]:2”
	//	- a domain name “example.com:1234”
	//	- a domain-socket address “/socket” “@socket”
	//	- empty string if domain is empty or AddrPort is invalid
	fmt.Stringer
}

// NewSocketAddressFromString returns SocketAddress based on
// address literal in string format or domain name
//   - network: [NetworkTCP]
//   - socketAddressString: “1.2.3.4:5” “[::]1” “example.com:3”
func NewSocketAddressFromString(
	network Network,
	socketAddressString string,
) (socketAddress SocketAddress) {
	var netipAddrPort netip.AddrPort
	var err error
	if netipAddrPort, err = netip.ParseAddrPort(socketAddressString); err != nil {
		// could not be parsed as address literal: must be domain name
		//	- “example.com”
		socketAddress = NewSocketAddress(network, socketAddressString)
		return
	}
	// has netip.AddrPort

	socketAddress = NewSocketAddressLiteral(network, netipAddrPort)

	return
}

// NewSocketAddressLiteral returns SocketAddress based on an IP address literal
//   - network: [NetworkTCP]
//   - addrPort: “1.2.3.4:5” “[::]1”
func NewSocketAddressLiteral(
	network Network,
	addrPort netip.AddrPort,
	fieldp ...*SaIPLiteral,
) (socketAddress SocketAddress) {

	// get a storage
	var a *SaIPLiteral
	if len(fieldp) > 0 {
		a = fieldp[0]
	}
	if a == nil {
		a = &SaIPLiteral{}
	}
	*a = SaIPLiteral{
		network:        network,
		addrPort:       addrPort,
		addrPortString: addrPort.String(),
	}
	socketAddress = a

	return
}

// NewSocketAddressLiteral returns SocketAddress based on an IP address literal
//   - network: [NetworkTCP]
//   - addrPort: “1.2.3.4:5” “[::]1”
func ParseSocketAddressFromNetAddr(
	netAddr net.Addr,
	fieldp ...*SaIPLiteral,
) (socketAddress SocketAddress, err error) {

	// get a storage
	var a *SaIPLiteral
	if len(fieldp) > 0 {
		a = fieldp[0]
	}
	if a == nil {
		a = &SaIPLiteral{}
	}

	if a.network, err = ParseNetwork(netAddr.Network()); perrors.IsPF(&err, "ParseNetwork %w", err) {
		return
	} else if a.addrPort, err = netip.ParseAddrPort(netAddr.String()); perrors.IsPF(&err, "ParseAddrPort %w", err) {
		return
	}
	a.addrPortString = a.addrPort.String()
	socketAddress = a

	return
}

// NewSocketAddress creates a SocketAddress based on
// domain name
//   - network: [NetworkTCP]
//   - domain: “example.com:5”
func NewSocketAddress(
	network Network,
	domain string,
	fieldp ...*SaDomain,
) (socketAddress SocketAddress) {

	var a *SaDomain
	if len(fieldp) > 0 {
		a = fieldp[0]
	}
	if a == nil {
		a = &SaDomain{}
	}
	*a = SaDomain{
		network: network,
		domain:  domain,
	}
	socketAddress = a

	return
}

// SaIPLiteral is [SocketAddress] implementation for
// address literal IPv4/IPv6
//   - all methods are fast non-converting
type SaIPLiteral struct {
	network        Network
	addrPort       netip.AddrPort
	addrPortString string
}

func (a SaIPLiteral) AddrPort() (addrPort netip.AddrPort) { return a.addrPort }
func (a SaIPLiteral) Network() (network Network)          { return a.network }
func (a SaIPLiteral) String() (s string)                  { return a.addrPortString }

// SaDomain is [SocketAddress] implementation for
// domain name or unix socket
//   - all methods are fast non-converting
//   - AddrPort returns nil, ie. invalid
type SaDomain struct {
	network Network
	domain  string
}

func (a SaDomain) AddrPort() (addrPort netip.AddrPort) { return }
func (a SaDomain) Network() (network Network)          { return a.network }
func (a SaDomain) String() (s string)                  { return a.domain }

type PreAllocSock struct {
	Literal SaIPLiteral
	Domain  SaDomain
}

// NetAddr exposes a SocketAddr as legacy
// [net.Addr] interface
//   - difference is Network method returns string
type NetAddr struct {
	SocketAddress
}

func (a *NetAddr) Network() (n string) {
	return a.SocketAddress.Network().String()
}

var _ net.Addr = &NetAddr{}
