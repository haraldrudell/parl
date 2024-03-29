/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"
	"net/netip"

	"github.com/haraldrudell/parl/perrors"
)

const Do46 = true

var IPv4DefaultNetwork = netip.MustParsePrefix("0.0.0.0/0")
var IPv6DefaultNetwork = netip.MustParsePrefix("::/0")

// MaskToBits returns number of leading 1-bits in byts
//   - convert from net.IPMask to netip.Prefix
func MaskToBits(byts []byte) (bits int, err error) {
	var hadZero bool
	for _, byt := range byts {
		if hadZero && byt != 0 {
			err = perrors.ErrorfPF("mask has intermediate zeroes: %v", byts)
			return
		} else if byt == 255 {

			// byte with all 1s
			bits += 8
			continue
		}

		// byte with mixed 0 and 1 bits
		hadZero = true
		for byt != 0 {
			if byt&128 != 0 {
				bits++
				byt <<= 1
				continue
			}

			// there was a zero bit before all 1 bits were found
			err = perrors.ErrorfPF("mask has intermediate zeroes: %v", byts)
			return
		}
	}
	return
}

// IPNetToPrefix returns the netip.Prefix that corresponds to older type net.IPNet
//   - net.IPNet input is "1.2.3.4/24" or "fe80::1/64"
//   - returned netip.Prefix values are valid
//   - returned IPv6 addresses has blank Zone
func IPNetToPrefix(netIPNet *net.IPNet, noIs4In6Translation ...bool) (prefix netip.Prefix, err error) {
	var i4Translation = true
	if len(noIs4In6Translation) > 0 {
		i4Translation = noIs4In6Translation[0]
	}
	var netipAddr netip.Addr
	var ok bool
	if netipAddr, ok = netip.AddrFromSlice(netIPNet.IP); !ok {
		// netIPNet.IP is []byte
		err = perrors.ErrorfPF("conversion to netip.Addr failed: IP: %#v", netIPNet.IP)
		return
	}
	// translate an IPv6 address that is 4in6 to IPv4
	//	- IPv6 "::ffff:127.0.0.1" becomes IPv4 "127.0.0.1"
	if i4Translation && netipAddr.Is4In6() {
		netipAddr = netip.AddrFrom4(netipAddr.As4())
	}
	var bits int
	if bits, err = MaskToBits(netIPNet.Mask); err != nil { // net.IPMask is []byte
		return
	}
	var p = netip.PrefixFrom(netipAddr, bits)
	if !p.IsValid() {
		err = perrors.ErrorfPF("conversion to netip.Addr failed net.IPNet: %#v", netIPNet.IP)
		return
	}

	prefix = p

	return
}

// AddrSlicetoPrefix returns a netip.Prefix slice from an Addr slice
//   - net.Interface.Addrs returns []net.Addr which is really []*net.IPNet
//   - IPNet is cidr not address: netip.Prefix
func AddrSlicetoPrefix(addrs []net.Addr, do46 bool) (prefixes []netip.Prefix, err error) {
	prefixes = make([]netip.Prefix, len(addrs))
	for i, netAddr := range addrs {
		var ipNet *net.IPNet
		var ok bool
		if ipNet, ok = netAddr.(*net.IPNet); !ok {
			err = perrors.ErrorfPF("not net.IPNet at #%d: %q", i, netAddr)
			return
		}
		var netipAddr netip.Addr
		if netipAddr, ok = netip.AddrFromSlice(ipNet.IP); !ok {
			err = perrors.ErrorfPF("AddrFromSlice at #%d: %q", i, netAddr)
			return
		}
		var bits int
		if bits, err = MaskToBits(ipNet.Mask); perrors.IsPF(&err, "AddrFromSlice at #%d: %q %q", i, netAddr, err) {
			return
		}
		if do46 && netipAddr.Is4In6() {
			netipAddr = Addr46(netipAddr)
			// bits is for IPv4 already
		}
		prefixes[i] = netip.PrefixFrom(netipAddr, bits)
	}
	return
}

func Prefix46(prefix netip.Prefix) (prefix46 netip.Prefix) {
	var addr = prefix.Addr()
	if !addr.Is4In6() {
		prefix46 = prefix
	} else {
		prefix46 = netip.PrefixFrom(Addr46(addr), prefix.Bits())
	}
	return
}
