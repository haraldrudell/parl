/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"errors"
	"net"
	"net/netip"

	"github.com/haraldrudell/parl/perrors"
)

var (
	// DefaultRouteIPv4 is the default route “0/0” for IPv4
	//   - [DefaultRouteIPv6]
	DefaultRouteIPv4 = netip.MustParsePrefix("0.0.0.0/0")
	// VPNRoute0IPv4 is overriding VPN route “0/1” for IPv4
	//   - [VPNRoute128IPv4] [VPNRouteIPv6]
	VPNRoute0IPv4 = netip.MustParsePrefix("0.0.0.0/1")
	// VPNRoute128IPv4 is overriding VPN route “128/1” for IPv4
	//   - [VPNRoute0IPv4] [VPNRouteIPv6]
	VPNRoute128IPv4 = netip.MustParsePrefix("128.0.0.0/1")
	// DefaultRouteIPv6 is the default route “::/0” for IPv6
	//   - [DefaultRouteIPv4]
	DefaultRouteIPv6 = netip.MustParsePrefix("::/0")
	// VPNRouteIPv6 is overriding VPN route “::/3” for IPv6
	//   - [VPNRoute0IPv4] [VPNRoute128IPv4]
	VPNRouteIPv6 = netip.MustParsePrefix("::/3")
	// LocalHost is hostname for locahost usable with both IPv4 and IPv6
	//	- routable without DNS
	LocalHost = "localhost"
)

// IsDirect determines if the route is direct
//   - route.Addr valid IPv4: isDirect true if prefix length is /32
//   - route.Addr valid IPv6: isDirect true if prefix length is /128
//   - route invalid: isDirect false
//   - —
//   - a direct route has mask 32 or 128 bit length /32 /128
//   - IsDirect considers an IPv4-mapped IPv6 address IPv6
//   - — therefore must have /128 prefix to be direct route
//   - with legacy type [net.IPMask], the mask has a prefix and a length
//   - — an IPv4-mapped IPv6 address is then paired with
//     a 32-bit-length IPv4 mask
//   - [netip.Prefix] does not have mask size
//   - — when legacy [net.IPNet] is converted to [netip.Prefix],
//     the IPv4-mapped IPv6 addresses is converted to IPv4
//   - — therefore the /128 approach is unlikely to be a problem
func IsDirect(route netip.Prefix) (isDirect bool) {
	return route.Addr().Is4() && route.Bits() == 32 ||
		route.Addr().Is6() && route.Bits() == 128
}

// the function returning prefix for [netip.Prefix]
var _ = netip.Prefix.Bits

// NetworkAddress returns the first address for a network
//   - addr: an IPv4 or IPv6 address defining the network
//   - routingPrefix: the number of significant bits forming network with remaining bits
//     used for host addressing
//   - network: the first network address
//   - — “fd19::1/64” → “fd19::/64”
//   - — “1.2.3.4/24” → “1.2.3.0/24”
//   - err: addr invalid, routingPrefix < 0,
//     routingPrefix > bit-lenght(addr) - 2
func NetworkAddress(addr netip.Addr, routingPrefix int) (network netip.Addr, err error) {

	// validate parameters
	if err = AddrRoutingPrefixValid(addr, routingPrefix); err != nil {
		return // invalid parameters: network is invalid
	}

	network = setEndingBits(addr, routingPrefix, setNetwork)

	return
}

// NetworkBroadcast returns the broadcast address for network
//   - addr: an IPv4 or IPv6 address defining the network
//   - routingPrefix: the number of significant bits separating the network
//     from host addressing
//   - — “fd19::1/64” → “fd19::ffff:ffff:ffff:ffff/64”
//   - — “1.2.3.4/24” → “1.2.3.255/24”
//   - err: addr invalid, routingPrefix < 0,
//     routingPrefix > bit-lenght(addr) - 2
func NetworkBroadcast(addr netip.Addr, routingPrefix int) (broadcast netip.Addr, err error) {

	// validate parameters
	if err = AddrRoutingPrefixValid(addr, routingPrefix); err != nil {
		return // invalid parameters: network is invalid
	}

	broadcast = setEndingBits(addr, routingPrefix, setBroadcast)

	return
}

// FirstHost returns the first host address for network
//   - addr: an IPv4 or IPv6 address defining the network
//   - routingPrefix: the number of significant bits separating the network
//     from host addressing
//   - — “fd19::1/64” → “fd19::1/64”
//   - — “1.2.3.4/24” → “1.2.3.1/24”
//   - err: addr invalid, routingPrefix < 0,
//     routingPrefix > bit-lenght(addr) - 2
func FirstHost(addr netip.Addr, routingPrefix int) (firstHost netip.Addr, err error) {

	// validate parameters
	if err = AddrRoutingPrefixValid(addr, routingPrefix); err != nil {
		return // invalid parameters: network is invalid
	}

	firstHost = setEndingBits(addr, routingPrefix, setFirstAddress)

	return
}

// AddrRoutingPrefixValid returns err nil if:
//   - addr is a valid IPv4 or IPv6 address
//   - routingPrefix is not negative
//   - routingPrefix is at least 2 less than [netip.Addr.Bitlen].
//     This means there are at least four host addresses available:
//   - host address 0: the network address
//   - 2 or more available host addresses
//   - host address all 1: the broadcast address
func AddrRoutingPrefixValid(addr netip.Addr, routingPrefix int) (err error) {

	// bitLen is 32 for IPv4 or 128 for IPv6
	//	- zero for invalid IP
	var bitLen = addr.BitLen()

	// addr invalid case
	if bitLen == 0 {
		err = perrors.NewPF("invalid netip.Addr")
		return
	}

	// routingPrefix invalid case
	//	- [netip.Prefix] has prefix -1 as invalid
	//	- legacy [net.IPMask] has prefix 0 as invalid
	//	- routing prefix zero is default route “0/0” or “::/0”
	if routingPrefix < 0 {
		err = perrors.ErrorfPF("invalid routingPrefix: %d", routingPrefix)
		return
	}

	// routingPrefix too large case
	//	- to use host addressing, 4 addresses are required
	//	- — the network address 0, the broadcast address -1 and
	//	- — at least two hosts
	//	- therefore maximum routing prefix is bitLen - 2 bits,
	//		2^2 is 4 host addresses
	if m := bitLen - hostAddresssingBits; routingPrefix > m {
		err = perrors.ErrorfPF("routingPrefix too large: %d > %d", routingPrefix, m)
		return
	}

	return
}

func setEndingBits(addr netip.Addr, routingPrefix int, addressChange setHostAdress) (newAddr netip.Addr) {

	// b is the value to set: 0 or 255
	var b byte
	if addressChange == setBroadcast {
		b = 255
	}

	// bitCount is the number of bits to modify at end
	//	- at least 2
	//	- max 32 for IPv4, 128 for Ipv6
	var bitCount = addr.BitLen() - routingPrefix

	// ipSlice is byte rpresentation of addr
	//	- length 4 for IPv4 and 16 for IPv6
	var ipSlice = addr.AsSlice()

	// index is current byte in ipSlice
	var index = len(ipSlice) - 1

	// modify entire bytes
	for bitCount > bitsPerByte {
		ipSlice[index] = b
		bitCount -= bitsPerByte
		index--
	}

	// modify the lowest bits of byte
	if bitCount > 0 {
		// if bitCount = 2, bitMask is 3
		var bitMask = byte(1<<bitCount - 1)
		// modify bits
		if b != 0 {
			// set bits to 1
			ipSlice[index] |= bitMask
		} else {
			// set the bits to zero
			ipSlice[index] &^= bitMask
		}
	}

	// first host case
	if addressChange == setFirstAddress {
		ipSlice[len(ipSlice)-1] |= 1
	}
	newAddr, _ /*ok*/ = netip.AddrFromSlice(ipSlice)

	return
}

// IsBroadcast determines whether addr is the last address of a routing prefix
//   - the last host address is typically broadcast
//   - addr: an IPv4 or IPv6 address defining the network
//   - routingPrefix: the number of significant bits separating the network
//     from host addressing
//   - isBroadcast true: if addr is the network broadcast address
//   - — “fd19::ffff:ffff:ffff:ffff/64” → true
//   - — “1.2.3.255/24” → true
//   - isBroadcast false: other address,
//     addr invalid, routingPrefix < 0,
//     routingPrefix > bit-lenght(addr) - 2
//   - —
//   - for 1.2.3.4/24 the network address 1.2.3.255 returns true
//   - an IP address can be split into a network and host addresses
//   - the network is defined by a routing prefix that has
//     a number of leading significant bits separating
//     the network from remaining bits used for host addressing
//   - the local broadcast address is the
//     all-ones host address of each network
//   - a cidr notes the significant bits: “1.2.3.4/24” means 24
//   - a subnet mask notes the same bits using leading 1’s: “255.255.255.0”
//   - [netip.Prefix.Bits] returns the routing prefix for that prefix or -1 if invalid
//   - [net.IPNet.Mask] returns the subnet mask for an IP network
//   - [net.IPMask.Size] returns the number of leading ones
//     or zero if invalid
func IsBroadcast(addr netip.Addr, routingPrefix int) (isBroadcast bool) {

	var broadcast, err = NetworkBroadcast(addr, routingPrefix)
	if err != nil {
		return // invalid parameters: isBroadcast false
	}

	// netip.Addr is comparable
	isBroadcast = addr == broadcast

	return
}

// IsErrClosed returns true if err is caused by socket closing while waiting for accept
//   - err: error returned by a net listener such as [net.Listener.Accept] [net.Conn.Accept]
//   - isErrNetClosing: true if error due to socket closed
func IsErrClosed(err error) (isErrNetClosing bool) {

	// opError indicates an error ocurring during the operation and
	// not a parameter error or such
	var opError *net.OpError

	// check for error during operation
	if !errors.As(err, &opError) {
		return // other error cause return: isErrNetClosing false
	}

	// check if cause is and it is that the listener was closed
	isErrNetClosing = opError.Err == net.ErrClosed

	return // isErrNetClosing valid
}

const (
	// the number of bits required to create host addressing
	hostAddresssingBits = 2
	// the number of bits per byte
	bitsPerByte = 8
)

const (
	// set the host address to the network address
	setNetwork setHostAdress = iota + 1
	// set the host address to the broadcast address
	setBroadcast
	// set the network’s first address, often a network server
	setFirstAddress
)

// setHostAdress controls how to modify host addressing
//   - [setNetwork] [setBroadcast] [setFirstAddress]
type setHostAdress uint8
