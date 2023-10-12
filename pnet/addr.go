/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"
	"net/netip"
	"strconv"

	"github.com/haraldrudell/parl/perrors"
)

// IsNonGlobalIPv6 returns if addr is an IPv6 address that should have zone
func IsNonGlobalIPv6(addr netip.Addr) (needsZone bool) {
	return addr.Is6() && (              //
	addr.IsInterfaceLocalMulticast() || // 0xffx1::
		addr.IsLinkLocalMulticast() || // 0xffx2::
		addr.IsLinkLocalUnicast() || // 0xfe8::
		addr.IsPrivate())
}

// EnsureZone adds IPv6 zone as applicable
//   - only non-global IPv6 should have zone
//   - if acceptNumeric true no index to interface-name translation attempts take place.
//     otherwise interface-name zone is preferred
//   - a non-numeric zone is attempted from: addr, ifName
//   - number cinversion to interfa e is attempted: addr, ifIndex
//   - acceptNumeric true leaves an existing numeric zone
func EnsureZone(addr netip.Addr, ifName string, ifIndex IfIndex, acceptNumeric ...bool) (addr2 netip.Addr, didChange, isNumeric bool, err error) {
	addr2 = addr
	var doNumeric bool
	if len(acceptNumeric) > 0 {
		doNumeric = acceptNumeric[0]
	}

	// does addr need zone?
	if !IsNonGlobalIPv6(addr) {
		return // this IPv6 address does not need zone return
	}

	// does addr already have non-numeric zone?
	var zone = addr.Zone()
	var ifi IfIndex
	var hasNumericZone bool
	if zone != "" {
		var number int
		var e error
		number, e = strconv.Atoi(zone)
		if e != nil {
			return // addr already has non-numeric zone
		}
		var ifiTry IfIndex
		ifiTry, e = NewIfIndexInt(number)
		if hasNumericZone = e == nil && ifiTry.IsValid(); hasNumericZone {
			if doNumeric {
				isNumeric = true
				return // zone is numeric and that should be used return
			}
			ifi = ifiTry
		}
	}

	//	- addr should have zone
	//	- addr has no zone or has numeric zone with doNumeric false

	// use ifName
	if didChange = ifName != ""; didChange {
		addr2 = netip.Addr.WithZone(addr, ifName)
		return
	}

	// use ifIndex if donumeric is true
	if doNumeric {
		if didChange = ifIndex.IsValid(); didChange {
			isNumeric = true
			addr2 = netip.Addr.WithZone(addr, ifIndex.String())
			return
		}
		err = perrors.NewPF("no zone in addr ifIndex ifName")
		return
	}

	// attempt translation of ifi ifIndex

	// translate addr numeric zone
	if hasNumericZone {
		var z string
		var isNo bool
		var e error
		if z, isNo, e = ifi.Zone(); z != "" && !isNo && e == nil {
			didChange = true
			addr2 = netip.Addr.WithZone(addr, z)
			return
		}
	}

	// translate ifIndex
	if ifIndex.IsValid() {
		var z string
		var isNo bool
		var e error
		if z, isNo, e = ifIndex.Zone(); z != "" && !isNo && e == nil {
			didChange = true
			addr2 = netip.Addr.WithZone(addr, z)
			return
		}
	}

	// no translation is available
	// doNumeric is false
	// fallback to any numeric

	// numeric addr.Zone
	if hasNumericZone {
		return // best is the numeric zone already in addr return
	}

	// numeric ifIndex
	if didChange = ifIndex.IsValid(); didChange {
		isNumeric = true
		addr2 = addr.WithZone(ifIndex.String())
		return
	}

	// it’s a failure
	err = perrors.NewPF("no successful translation or zone in addr ifIndex ifName")

	return
}

//var isDigits = regexp.MustCompile(`^[0-9]+$`).MatchString

// Zone examines the zone included in addr
//   - no zone: hasZone, isNumeric false
//   - numeric zone "1": hasZone true, isNumeric false
//   - interface-name zone "eth0": hasZone, isNumeric true
func Zone(addr netip.Addr) (zone string, znum int, hasZone, isNumeric bool) {
	zone = addr.Zone()
	if hasZone = zone != ""; !hasZone {
		return
	}
	var err error
	znum, err = strconv.Atoi(zone)
	isNumeric = err == nil
	return
}

// IPAddr returns IPAddr from IP and IfIndex to IPAddr
func IPAddr(IP net.IP, index IfIndex, zone string) (ipa *net.IPAddr, err error) {
	ipa = &net.IPAddr{IP: IP}
	if IsIPv6(IP) {
		if zone != "" {
			ipa.Zone = zone
		} else {
			ipa.Zone, _, err = index.Zone()
		}
	}
	return
}

// AddrToIPAddr: Network() "ip", no port number
func AddrToIPAddr(addr netip.Addr) (addrInterface net.Addr) {
	if !addr.IsValid() {
		panic(perrors.NewPF("invalid netip.Addr"))
	}
	return &net.IPAddr{IP: addr.AsSlice(), Zone: addr.Zone()}
}

// Addr46 convert 4in6 to 4 for consistent IPv4/IPv6
func Addr46(addr netip.Addr) (addr46 netip.Addr) {
	if addr.Is4In6() {
		addr46 = netip.AddrFrom4(addr.As4())
	} else {
		addr46 = addr
	}
	return
}
