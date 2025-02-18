/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
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
//   - route Is4: IPv4: true if prefix length is /32
//   - — IPv4-mapped IPv6 addresses are considered IPv6
//   - route.Is6: IPv6: true if prefix length is /128
//   - route invalid: isDirect false
//   - a direct route has mask 32 or 128 bit length /32 /128
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
//   - network invalid: addr or routingPrefix is invalid
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
func FirstHost(addr netip.Addr, routingPrefix int) (firstHost netip.Addr, err error) {

	// validate parameters
	if err = AddrRoutingPrefixValid(addr, routingPrefix); err != nil {
		return // invalid parameters: network is invalid
	}

	firstHost = setEndingBits(addr, routingPrefix, setFirstAddress)

	return
}

// AddrRoutingPrefixValid returns true if:
//   - addr is a valid IPv4 or IPv6 address
//   - routingPrefix is valid and not zero or negative
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
	if routingPrefix <= 0 {
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
//   - for 1.2.3.4/24 the network address 1.2.3.255 returns true
//     -
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

// IsErrClosed returns true if err is when waiting for accept and the socket is closed
//   - can be used with net.Conn.Accept
func IsErrClosed(err error) (isErrNetClosing bool) {
	// if err is nil, ok is false
	if netOpError, ok := err.(*net.OpError); ok { // error occured during the operation
		isErrNetClosing = netOpError.Err == net.ErrClosed // and it is that the listener was closed
	}
	return
}

func NetworkPrefixBitCount(byts []byte) (bits int) {

	// count bits that are 1 from the high order bit until a zero bit is found
	for _, byt := range byts {
		if byt == 255 {
			bits += 8
			continue
		}
		for byt != 0 {
			if byt&128 != 0 {
				bits++
			}
			byt <<= 1
		}
		break
	}
	return
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
