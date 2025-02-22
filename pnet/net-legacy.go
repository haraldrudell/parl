/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"
	"net/netip"
	"strconv"
	"strings"

	"github.com/haraldrudell/parl/perrors"
)

// TODO 250217 extract to separate package pnetx
//	- but uses pnet: Addr46 IfIndex

// legacy [net.Addr] is [netip.Addr]
//   - Interface [netip.Addr.Network] [netip.Addr.String]
//   - possible [net.Addr.Network]: "tcp" "tcp4" "tcp6"
//     "udp" "udp4" "udp6" "ip" "ip4" "ip6" "unix" "unixgram" "unixpacket"
//   - interface implemented by:
//   - [net.IPNet] "ip+net" “1.2.3.4/24” or “::/3”
//   - [net.UDPAddr] “udp”
//   - [net.TCPAddr] “tcp”
//   - [net.IPAddr] "ip"
//   - [net.UnixAddr] "unix" "unixgram" "unixpacket"
var _ net.Addr

// legacy [net.UDPAddr] is [netip.AddrPort]
var _ net.UDPAddr

// legacy [net.TCPAddr] is [netip.AddrPort]
var _ net.TCPAddr

// legacy [net.UnixAddr]
var _ net.UnixAddr

// legacy [pnet.IP] is [netip.Addr]
var _ net.IP

// legacy [net.IPMask] is [netip.Prefix.Bits]
//   - has no new-function
var _ net.IPMask

// the function returning prefix for legacy [net.IPMask]
var _ = net.IPMask.Size

// legacy [net.IPAddr] is [netip.Addr]
//   - is IP address holding IPv6 Zone information
var _ net.IPAddr

// legacy [net.IPNet] is [netip.Prefix]
var _ net.IPNet

//new type directory:

// [netip.Addr] is legacy [net.Addr] [net.IPAddr] [net.IP]
var _ netip.Addr

// [netip.AddrPort] is legacy [net.UDPAddr] [net.TCPAddr]
var _ netip.AddrPort

// [netip.Prefix] is legacy [net.IPMask] [net.IPNet]
var _ netip.Prefix

var (
	// IPv4loopback is [net.IP] for localhost IPv4
	//   - similar to [net.IPv6loopback] 16-byte “::1”
	//   - legacy net.IP type for use with [x509.Certificate]
	IPv4loopback net.IP = net.IPv4(127, 0, 0, 1)
)

// AddrToIPAddr returns [net.Addr] interface IP string-address for a
// [netip.Addr] address literal
//   - addr: newer [netIP.Addr] IPv4/IPv6 address literal
//   - — panic if invalid
//   - addrInterface: legacy interface returning strings
//   - — Network [net.Addr.Network] returns “ip”, ie. no port number
//   - — String [net.IPAddr.String] returns “1.2.3.4” or “fe80::%eth0”
//   - — strings are used to support DNS names
//   - — implementation is [*net.IPAddr]
//   - —
//   - [net.Addr] is legacy interface for [net.Dial] using strings
//   - — [net.Addr.Network] is string like “tcp”
//   - — [net.Addr.String] is socket address literal like “1.2.3.4:80”
//   - TODO 250217 unused, deprecate
//
// legacy net pre-go1.18 220315 functions:
//   - [AddrToIPAddr] returns legacy [net.Addr] string IP address from [netip.Addr]
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
func AddrToIPAddr(addr netip.Addr) (addrInterface net.Addr) {
	if !addr.IsValid() {
		panic(perrors.NewPF("invalid netip.Addr"))
	}
	addrInterface = &net.IPAddr{IP: addr.AsSlice(), Zone: addr.Zone()}

	return
}

// AddrPortToUDPAddr returns [net.Addr] udp string socket address for a
// [netip.AddrPort] socket literal
//   - —
//   - legacy [*net.UDPAddr] is used by [net.ListenUDP] [net.Dialer.LocalAddr]
//   - —
//
// legacy net pre-go1.18 220315 functions:
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
func AddrPortToUDPAddr(addrPort netip.AddrPort) (addrInterface net.Addr) {
	var IP, port, zone = SplitAddrPort(addrPort)
	addrInterface = &net.UDPAddr{IP: IP, Port: port, Zone: zone}

	return
}

// IPAddr returns [*net.IPAddr] from [net.IP] [IfIndex] and zone
//   - TODO 250217 deprecate, unused
//
// legacy net pre-go1.18 220315 functions:
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
func IPAddr(IP net.IP, index IfIndex, zone string) (ipa *net.IPAddr, err error) {

	// populate IP address
	ipa = &net.IPAddr{IP: IP}

	// add optional zone
	if IsIPv6(IP) {
		if zone != "" {
			ipa.Zone = zone
		} else {
			ipa.Zone, _, err = index.Zone()
		}
	}

	return
}

// AddrPortToUDPAddr2 returns legacy [*net.UDPAddr] from [netip.AddrPort]
//   - —
//   - — [*net.UDPAddr] is used by [net.ListenUDP] [net.Dialer.LocalAddr]
//
// legacy net pre-go1.18 220315 functions:
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
func AddrPortToUDPAddr2(addrPort netip.AddrPort) (addr *net.UDPAddr) {
	var IP, port, zone = SplitAddrPort(addrPort)
	addr = &net.UDPAddr{IP: IP, Port: port, Zone: zone}

	return
}

// SplitAddrPort converts [netip.AddrPort] to legacy types IP, port and zone
//   - IP: legacy [net.IP] IPv4/IPv6 byte-slice address
//   - port: port number
//   - zone: interfac e name or interface index numeric string
//     optional for IPv6 addresses
//
// legacy net pre-go1.18 220315 functions:
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
func SplitAddrPort(addrPort netip.AddrPort) (IP net.IP, port int, zone string) {
	if !addrPort.IsValid() {
		panic(perrors.NewPF("invalid netip.AddrPort"))
	}
	IP = addrPort.Addr().AsSlice()
	port = int(addrPort.Port())
	zone = addrPort.Addr().Zone()
	return
}

// AddrPortToTCPAddr returns [net.Addr] interface "tcp" [*net.TCPAddr] from [netip.AddrPort]
//   - —
//   - [net.TCPAddr] is used by [net.Dialer.LocalAddr]
//
// legacy net pre-go1.18 220315 functions:
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
func AddrPortToTCPAddr(addrPort netip.AddrPort) (addrInterface net.Addr) {
	IP, port, zone := SplitAddrPort(addrPort)
	return &net.TCPAddr{IP: IP, Port: port, Zone: zone}
}

// InvertMask returns inverts legacy [net.IPMask]
//   - the mask for “1.2.3.4/24” is normally ffffff00 or []byte{255, 255, 255, 0}
//   - —
//   - TODO 250217 unused possibly deprecate
//
// legacy net pre-go1.18 220315 functions:
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
func InvertMask(IPMask net.IPMask) (out net.IPMask) {
	out = make(net.IPMask, len(IPMask))
	for i, b := range IPMask {
		out[i] = ^b
	}
	return
}

// IPNetString returns abbreviated string from “0/0” from legacy [net.IPNet]
//   - TODO 250217 unsued possibly deprecate
//
// legacy net pre-go1.18 220315 functions:
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
func IPNetString(ipNet net.IPNet) (s string) {

	// the /24 or /32 of CIDR
	var ones, _ = ipNet.Mask.Size()
	s = shorten(ipNet.IP) + "/" + strconv.Itoa(ones)

	return
}

// shorten shortens a string IPv4 address to “127” from legacy [net.IP]
func shorten(IP net.IP) (s string) {

	// convert to string
	s = IP.String()
	if len(IP) != net.IPv4len {
		return // do not shorten IPv6
	}

	// remove ending “.0”
	for strings.HasSuffix(s, zeroSuffix) {
		s = s[:len(s)-len(zeroSuffix)]
	}

	return
}

// IsIPv4 returns true if legacy [net.IP] is IPv4 or IPv4-mapped IPv6 address and not unset, corrupt or IPv6
//   - ip: a legacy IP address to examine
//   - isIPv4 true: ip is IPv4 or IPv4-mapped IPv6 address
//   - — IPv4-mapped IPv6 address are considered IPv4 “::ffff:1.2.3.4”
//   - isIPv4 false: ip is uninitialized, corrupt or other IPv6
//     -
//   - IP implementation is []byte byte-slice
//   - an unitialized net.IP is nil
//   - net.IP of length other than 0, 4 or 16 is invalid
//
// legacy net pre-go1.18 220315 functions:
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
func IsIPv4(ip net.IP) (isIPv4 bool) {

	// [net.IP.To4] returns n IPv4 address for valid IPv4 or IPv4-mapped IPv6 address
	//	- nil otherwise
	isIPv4 = len(ip.To4()) == net.IPv4len

	return
}

// IsIPv6 returns true if legacy [net.IP] is IPv6 and not unset or IPv4 or IPv4 in IPv6
//   - ip: an IPv4/IPv6 address to examine
//   - isIPv6 true: ip is valid IPv6 but not IPv4-mapped IPv6 address
//   - — IPv4-mapped addresses are considered IPv4 “::ffff:1.2.3.4”
//   - isIPv6 false: ip is nil, invalid or IPv4 or IPv4-mapped IPv6 address
//     -
//   - IP implementation is []byte byte-slice
//   - an unitialized net.IP is nil
//   - net.IP of length other than 0, 4 or 16 is invalid
//
// legacy net pre-go1.18 220315 functions:
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
func IsIPv6(ip net.IP) (isIPv6 bool) {

	// [net.IP.To4] returns n IPv4 address for valid IPv4 or IPv4-mapped IPv6 address
	//	- nil otherwise
	isIPv6 = len(ip.To4()) != net.IPv4len && len(ip) == net.IPv6len

	return
}

// IsValid returns true if legacy [net.IP] is an initialized IPv4 or IPv6 address
//   - ip: an IP address
//   - isValid true: ip is initialized IPv4 or IPv6 address
//
// legacy net pre-go1.18 220315 functions:
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
func IsValid(ip net.IP) (isValid bool) {
	isValid =
		len(ip) == net.IPv4len ||
			len(ip) == net.IPv6len
	return
}

// IsNzIP returns true if legacy [net.IP] is valid IPv4 or IPv6 that is not the zero address
//   - ip is IPv4 other than “0.0.0.0”: isNzIP true
//   - — including IPv4-mapped addresses “::ffff:1.2.3.4”
//   - ip is IPv6 other than “::”: isNzIP true
//   - ip unintialized or bad: isNzIP false
//
// legacy net pre-go1.18 220315 functions:
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
func IsNzIP(ip net.IP) (isNzIP bool) {

	// IsValid is true if ip is initialized IPv4 or IPv6
	if IsValid(ip) {
		// [net.IP.IsUnspecified] checks against “0/0” and “::/0”
		isNzIP = !ip.IsUnspecified()
	}

	return
}

// IPNetToPrefix returns [netip.Prefix] for legacy [*net.IPNet]
//   - net.IPNet: legacy network prefix string like "1.2.3.4/24" or "fe80::1/64"
//   - noIs4In6Translation missing: default is to translate IPv4 embedded in
//     IPv6 to an IPv4 prefix address
//   - noIs4In6Translation Dp46No: IPv4 translation
//   - prefix: valid returned prefix
//   - err: bad [net.IPNet.IP], bad [net.IPNet.Mask], mask does not fit address
//   - returned IPv6 addresses has blank Zone
//
// legacy net pre-go1.18 220315 functions:
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
func IPNetToPrefix(netIPNet *net.IPNet) (prefix netip.Prefix, err error) {

	// get network address from legacy [net.IP]
	var netipAddr netip.Addr
	var ok bool
	// [net.IPNet.IP] is []byte
	if netipAddr, ok = netip.AddrFromSlice(netIPNet.IP); !ok {
		// must be length 4 or 16 or error
		err = perrors.ErrorfPF("conversion to netip.Addr failed: IP: %#v", netIPNet.IP)
		return // [netIPNet.IP] bad length error return
	}

	// get prefix bits: 0–32 for IPv4, 0–128 for IPv6
	var bits int
	// true if the mask is 128-bit IPv6
	//	- false if mask is 32-bit IPv4
	//	- any other length is error
	var isIPv6 bool
	if bits, isIPv6, err = MaskToBits(netIPNet.Mask); perrors.IsPF(&err, "%w mask: %v", err, netIPNet.Mask) {
		return
	}

	// do possible IPv4 in IPv6 translation
	if netipAddr.Is4In6() {
		netipAddr = Addr46(netipAddr)
		if isIPv6 {
			// convert mask to IPv4
			isIPv6 = false
			// for IPv6 mask less than /96, IPv4 mask is zero
			if bits < 128-32 {
				bits = 0
			} else {
				// IPv6 /128 → IPv4 /32
				// IPv6 /96 → IPv4 /0
				bits = 32 - (128 - bits)
			}
		}
	}

	// ensure Addr and mask address family matches
	if netipAddr.Is6() {
		if !isIPv6 {
			err = perrors.ErrorfPF("IPv6 address with IPv4 mask")
			return
		}
	} else if isIPv6 {
		err = perrors.ErrorfPF("IPv4 address with IPv6 mask")
		return
	}

	// create [netip.Prefix]
	var p = netip.PrefixFrom(netipAddr, bits)
	if !p.IsValid() {
		// only if netipAddr invalid or bits negative or too large
		err = perrors.ErrorfPF("conversion to netip.Addr failed net.IPNet: %#v", netIPNet.IP)
		return // mismatched IP address familty and prefix bits error return
	}
	prefix = p

	return
}

// AddrSlicetoPrefix returns a [netip.Prefix] list from an [net.Addr] list
//   - converts the result from [net.Interface.Addrs] to non-legacy types
//   - addrs: list of [net.Addr], ie. [*net.IPNet] listing cidr “1.2.3.4/24” or “::/3”
//   - err: element not *net.IPNet, bad ipNet.IP, bad ipNet.Mask slice,
//     mismatched address family between addr and mask
//   - —
//   - IPv4-mapped IPv6 address is returned as IPv4
//   - — any IPv6 mask is then converted to IPv4 mask
//   - IPv4-mapped IPv6 address: “::ffff:127.0.0.1/8” often has an IPv4 mask
//   - — [net.IPNet.IP] is IPv6: “::ffff:127.0.0.1”
//   - — [net.IPNet.Mask] is 32-bit IPv4: [255, 0, 0, 0]
//   - [net.Interface.Addrs] returns []net.Addr which is really []*net.IPNet
//
// legacy net pre-go1.18 220315 functions:
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
func AddrSlicetoPrefix(addrs []net.Addr) (prefixes []netip.Prefix, err error) {

	// iterate of input
	var ps = make([]netip.Prefix, len(addrs))
	for i, netAddr := range addrs {

		// type assert net.Addr to *net.IPNet
		var ipNet *net.IPNet
		var ok bool
		if ipNet, ok = netAddr.(*net.IPNet); !ok {
			err = perrors.ErrorfPF("not net.IPNet at #%d: %q", i, netAddr)
			return // [net.Addr] not [*net.IPNet] error return
		}

		var p netip.Prefix
		if p, err = IPNetToPrefix(ipNet); err != nil {
			err = perrors.ErrorfPF("AddrFromSlice at #%d: %w", i, err)
			return
		}
		ps[i] = p
	}
	prefixes = ps

	return
}

// MaskToBits validates legacy [net.IPMask] and returns number of leading 1-bits
//   - ones: the network prefix length in bits 0–128: “1.2.3.4/24”: 24, “::/3”: 3
//   - isIPv6 true: the mask length is for IPv6: 128 bits, otherwise IPv4: 32 bits
//   - err any error condition:
//   - — uninitialized: mask is nil or length zero
//   - — bad length: mask byte-length does not match IPv4: 4 bytes 32 bits or
//     IPv6: 16 bytes 128 bits
//   - — corrupt: mask is not leading all-one bits, zero or more, followed by all-zero bits
//   - MaskToBits is used to create [netip.Prefix] from legacy [net.IPMask]
//
// legacy net pre-go1.18 220315 functions:
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
func MaskToBits(mask net.IPMask) (ones int, isIPv6 bool, err error) {

	// length of mask 0–128, unit bits
	var bits int
	ones, bits = mask.Size()

	// for illegal mask not strictly ones followed by zeroes: [net.IPMask.Size] returns 0, 0
	//	- Size does not check mask length to match IPv4 or IPv6
	//	- Size does not explicitly check for nil or zero-length mask
	switch len(mask) {
	case 0: // uninitialized or invalid mask
		if len(mask) == 0 {
			err = perrors.NewPF("uninitialized mask nil or zero length")
			return
		}
		err = perrors.ErrorfPF("mask has intermediate zeroes: %v", mask)
	case net.IPv4len: // valid IPv4 mask
	case net.IPv6len: // valid IPv6 mask
		isIPv6 = true
	default: // mask of bad length
		err = perrors.ErrorfPF("invalid mask length: %d allowed: IPv4: %d bytes; IPv6: %d bytes",
			bits, net.IPv4len, net.IPv6len,
		)
	}

	return
}

const (
	// zeroSuffix is used to shorten IPv4 addresses: “0.0.0.0/1” → “0/1”
	zeroSuffix = ".0"
)
