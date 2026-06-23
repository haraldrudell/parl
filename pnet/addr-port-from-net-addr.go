// © 2026–present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
// All rights reserved

package pnet

import (
	"net"
	"net/netip"

	"github.com/haraldrudell/parl/perrors"
)

// AddrPortFromAddr returns [netip.AddrPort] address literal from
// legacy [net.Addr] implemented by [*net.TCPAddr] address literal
//   - addr: [net.Addr] interface for tcp-network address
//   - — implemented by [*net.TCPAddr] legacy IP address with zone
//   - near: valid [netip.AddrPort] binary-coded socket address
//     “1.2.3.4:80” “::1:443” with optional zone
//   - err: concrete type not [*net.TCPAddr], bad [net.IP] length
//   - —
//   - — port number is not checked for being uint16
//     -— zone is not validated
//   - [net.Addr] is legacy type [net.Dial] uses to enable DNS strings for
//     socket address
//   - [*net.TCPAddr] returns
//   - — [net.TCPAddr.Network] “tcp” [NetworkTCP]
//   - — [net.TCPAddr.String] “[fe80::%eth0]:80”
//
// legacy net pre-go1.18 220315 functions:
//   - [AddrPortFromAddr] returns [netip.AddrPort] address literal from
//     legacy [net.Addr] implemented by [*net.TCPAddr]
//   - [AddrToIPAddr] returns [net.Addr] string IP address from [netip.Addr]
//   - [AddrPortToTCPAddr] returns legacy “tcp” [net.Addr] interface string socket address [*net.TCPAddr] from [netip.AddrPort]
//   - [AddrPortToUDPAddr] returns legacy “udp” [net.Addr] interface string socket address [*net.UDPAddr] from [netip.AddrPort]
//   - [AddrPortToUDPAddr2] returns legacy “udp” [*net.UDPAddr] string socket address from [netip.AddrPort]
//   - [AddrSlicetoPrefix] returns a [netip.Prefix] list from legacy [net.Addr] list
//   - [InvertMask] inverts legacy [net.IPMask]
//   - [IPAddr] returns legacy “ip” [*net.IPAddr] interface string socket address [*net.IPAddr] from legacy [net.IP] [IfIndex] and zone
//   - [IPNetString] returns abbreviated IPv4 “0/0” from legacy [net.IPNet]
//   - [IPNetToPrefix] returns [netip.Prefix] for legacy [*net.IPNet]
//   - [IsIPv4] returns true if legacy [net.IP] is IPv4 or IPv4 in IPv6 and not unset or IPv6
//   - [IsIPv6] returns true if legacy [net.IP] is IPv6 and not unset or IPv4 or IPv4 in IPv6
//   - [IsNzIP] returns true if legacy [net.IP] is valid IPv4 or IPv6 that is not the zero address]
//   - [IsValid] returns true if legacy [net.IP] is an initialized IPv4 or IPv6 address]
//   - [MaskToBits] returns [netip.Prefix.Bits] from legacy [net.IPMask]
//   - [SplitAddrPort] returns legacy [net.IP], port and zone from [netip.AddrPort]
func AddrPortFromAddr(netAddr net.Addr) (addrPort netip.AddrPort, err error) {
	var ip netip.Addr
	var ok bool
	var zone string
	var port uint16
	switch v := netAddr.(type) {
	case *net.TCPAddr:
		ip, ok = netip.AddrFromSlice(v.IP)
		zone = v.Zone
		port = uint16(v.Port)
	case *net.UDPAddr:
		ip, ok = netip.AddrFromSlice(v.IP)
		zone = v.Zone
		port = uint16(v.Port)
	default:
		err = perrors.ErrorfPF("Bad address type %T expected *net.TCPAddr or *net.UDPAddr",
			netAddr,
		)
		return
	}
	if !ok {
		err = perrors.ErrorfPF("Bad net.Addr: %T %s",
			netAddr, netAddr.String(),
		)
		return
	} else if zone != "" {
		ip = ip.WithZone(zone)
	}
	addrPort = netip.AddrPortFrom(ip, port)

	return
}
