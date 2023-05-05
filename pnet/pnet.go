/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package pnet provides IP-related functions with few dependencies beyond the net package
package pnet

import (
	"net"
	"net/netip"
	"strconv"
	"strings"
)

const (
	// DefaultRouteIPv4 is the default route 0/0 for IPv4
	DefaultRouteIPv4 = "0.0.0.0/0"
	// VPNRoute0IPv4 is overriding VPN route 0/1 for IPv4
	VPNRoute0IPv4 = "0.0.0.0/1"
	// VPNRoute128IPv4 is overriding VPN route 128/1 for IPv4
	VPNRoute128IPv4 = "128.0.0.0/1"
	// DefaultRouteIPv6 is the default route ::/0 for IPv6
	DefaultRouteIPv6 = "::/0"
	// VPNRouteIPv6 is overriding VPN route ::/3 for IPv6
	VPNRouteIPv6 = "::/3"
	zeroSuffix   = ".0"
)

// IsNetwork determines if IP is the network address (all zeros) for this Mask
// for 1.2.3.4/24 the network address 1.2.3.0 returns true
func IsNetwork(IP net.IP, IPMask net.IPMask) (isNetwork bool) {
	if len(IP) != net.IPv4len && len(IP) != net.IPv6len {
		return
	}
	isNetwork = IsZeros(IP.Mask(InvertMask(IPMask)))
	return
}

// IsZeros determines if every byte of the IP address is zero
func IsZeros(p net.IP) bool {
	for i := 0; i < len(p); i++ {
		if p[i] != 0 {
			return false
		}
	}
	return true
}

// IsDirect determines if the route is direct
//   - a direct route has mask 32 or 128 bit length /32 /128
func IsDirect(route netip.Prefix) bool {
	return route.Addr().Is4() && route.Bits() == 32 ||
		route.Addr().Is6() && route.Bits() == 128
}

// IsIPv4 determines if net.IP value is IPv4
func IsIPv4(ip net.IP) (isIPv4 bool) {
	isIPv4 = len(ip.To4()) == net.IPv4len
	return
}

// IsIPv6 determines if net.IP value is IPv6
func IsIPv6(ip net.IP) (isIPv6 bool) {
	isIPv6 = len(ip.To4()) != net.IPv4len && len(ip) == net.IPv6len
	return
}

// IsNzIP is ip set and not zero
func IsNzIP(ip net.IP) bool {
	return ip != nil && !ip.IsUnspecified()
}

// IsBroadcast determines whether addr is the last address for Mask
//   - the last address is typically broadcast
//   - for 1.2.3.4/24 the network address 1.2.3.255 returns true
func IsBroadcast(addr netip.Addr, IPMask net.IPMask) (isBroadcast bool) {
	if !addr.IsValid() {
		return
	}

	// convert to net,.IP to use Mask
	var netIP = net.IP(addr.AsSlice())

	invertedMask := InvertMask(IPMask)
	isBroadcast = netIP.Mask(invertedMask).Equal(net.IP(invertedMask))
	return
}

// InvertMask inverts the bits of a mask
// the mask for 1.2.3.4/24 is normally ffffff00 or []byte{255, 255, 255, 0}
func InvertMask(IPMask net.IPMask) (out net.IPMask) {
	out = make(net.IPMask, len(IPMask))
	for i, b := range IPMask {
		out[i] = ^b
	}
	return
}

// IPNetString is abbreviated form 0/0
func IPNetString(ipNet net.IPNet) (s string) {
	ones, _ := ipNet.Mask.Size() // the /24 or /32 of CIDR
	s = shorten(ipNet.IP) + "/" + strconv.Itoa(ones)
	return
}

func shorten(IP net.IP) (s string) {
	s = IP.String()
	if len(IP) != net.IPv4len {
		return
	}
	for strings.HasSuffix(s, zeroSuffix) {
		s = s[:len(s)-len(zeroSuffix)]
	}
	return
}

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
