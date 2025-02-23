/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net/netip"
	"strconv"

	"github.com/haraldrudell/parl/perrors"
)

// IsNonGlobalIPv6 returns whether addr is IPv6 address that
// should have zone
//   - addr: valid IPv6
//   - needsZone true: addr is valid IPv6 and either:
//   - — interface-local multicast
//   - — link-local multicast
//   - — link-local unicast
//   - — private, ie. non-public, IPv6 address
func IsNonGlobalIPv6(addr netip.Addr) (needsZone bool) {
	return addr.Is6() && (              //
	addr.IsInterfaceLocalMulticast() || // 0xffx1::
		addr.IsLinkLocalMulticast() || // 0xffx2::
		addr.IsLinkLocalUnicast() || // 0xfe8::
		addr.IsPrivate())
}

const (
	// [EnsureZone] ifName missing
	NoIfName = ""
	// [EnsureZone] ifName missing
	NoIfIndex = IfIndex(0)
)

// EnsureZone adds IPv6 zone as applicable
//   - only non-global IPv6 should have zone (routing table manipulation)
//   - addr: an IP address including invalid
//   - ifName: network interface name for non-numeric zone “lo” “eth0”.
//     [NoIfName] to allow numeric index
//   - ifIndex: network interface index for numeric-zone: 1…
//     [NoIfIndex] for no value
//   - acceptNumeric missing: index to interface-name translation is attemped
//   - acceptNumeric [ZoneNumericYes]: any numeric zone interface index in addr is accepted.
//     isNumeric is true if addr2 has numeric zone
//   - addr2: address that if IPv6 may have zone added
//   - — if addr isn’t non-public IPv6 or invalid, addr2 is addr
//   - — if addr has non-numeric addr.Zone, addr2 is addr
//   - — if addr has numeric zone and acceptNumeric: [ZoneNumericYes],
//     addr2 is addr
//   - — if addr is non-public without zone:
//   - — — if ifName present, addr2 has ifName as zone
//   - — — if ifIndex valid and acceptNumeric [ZoneNumericYes],
//     addr2 has ifIndex as zone
//   - — — if numeric addr zone can be ranslated to non-numeric,
//     that is the zone
//   - — — if ifIndex can be translated to non-numeric zone,
//     that is the zone
//   - didChange true: zone was added
//   - isNumeric true: a numeric zone was added
//   - err: zone was required and
//   - — addr did not have non-numeric zone
//   - — acceptNumeric [ZoneNumericYes] but addr.zone was not numeric and
//     ifIndex was invalid
//   - — ifName zone was not present
//   - — translation of numeric addr.zone failed
//   - — translation of ifIndex failed
//   - — no numeric addr.zone was available
//   - — ifIndex was invalid
func EnsureZone(addr netip.Addr, ifName string, ifIndex IfIndex, acceptNumeric ...ZoneArg) (addr2 netip.Addr, didChange, isNumeric bool, err error) {
	addr2 = addr

	// does addr need zone?
	if !IsNonGlobalIPv6(addr) {
		return // this IPv6 address does not need zone return
	}

	// doNumeric default false
	var doNumeric = len(acceptNumeric) > 0 && acceptNumeric[0] == ZoneNumericYes
	// whether addr already had numeric zone?
	var hasNumericZone bool
	// the zone initially in addr
	var inputZone = addr.Zone()
	// any interface index already present in addr zone
	var inputZoneIndex IfIndex

	// examine any zone in addr
	if inputZone != "" {

		// check for addr input zone non-numeric
		// addr possible input numeric zone
		var number int
		var e error
		number, e = strconv.Atoi(inputZone)
		if e != nil {
			return // addr already has non-numeric zone
		}
		// number is valid addr input numeric index

		var ifiTry IfIndex
		ifiTry, e = NewIfIndexInt(number)
		if hasNumericZone = e == nil && ifiTry.IsValid(); hasNumericZone {
			if doNumeric {
				isNumeric = true
				return // zone is numeric and that should be used return
			}
			inputZoneIndex = ifiTry
		}
	}
	//	- addr should have zone and either:
	//	- addr has no zone or
	//	- addr has numeric zone with doNumeric false

	// use ifName as zone if present
	if didChange = ifName != ""; didChange {
		addr2 = netip.Addr.WithZone(addr, ifName)
		return // ifName zone added return
	}

	// use ifIndex if donumeric is true
	if doNumeric {

		// use valid ifIndex as zone
		if didChange = ifIndex.IsValid(); didChange {
			isNumeric = true
			addr2 = netip.Addr.WithZone(addr, ifIndex.String())
			return // ifIndex as zone return
		}

		// should have numeric index but
		//	- addr has no numeric zone
		//	- ifIndex is invalid
		err = perrors.NewPF("no zone in addr ifIndex ifName")
		return // no nuemric zone value error return
	}
	// addr should have non-numeric zone and either:
	//	- has no zone or
	//	- has numeric zone

	// attempt translation of addr input numeric zone to non-numeric zone
	if hasNumericZone {
		var ifiZone string
		var ifiZoneIsNumeric bool
		var e error
		if ifiZone, ifiZoneIsNumeric, e = inputZoneIndex.Zone(); ifiZone != "" && !ifiZoneIsNumeric && e == nil {
			didChange = true
			addr2 = netip.Addr.WithZone(addr, ifiZone)
			return // addr input numeric zone translated to non-numeric zone return
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
			return // ifIndex translated to non-numeric zone return
		}
	}
	// addr should have non-numeric zone and:
	//	- addr has no zone or
	//	- addr has no translatable numeric zone and
	//	- ifindex is not translatable

	// numeric addr.Zone
	if hasNumericZone {
		return // best is the numeric zone already in addr return
	}

	// numeric ifIndex
	if didChange = ifIndex.IsValid(); didChange {
		isNumeric = true
		addr2 = addr.WithZone(ifIndex.String())
		return // numeric ifIndex zone return
	}

	// it’s a failure
	err = perrors.NewPF("no successful translation or zone in addr ifIndex ifName")

	return
}

// Zone examines the zone included in addr
//   - no zone: hasZone: false, isNumeric: false
//   - numeric zone “1”: hasZone: true, isNumeric: true.
//     znum is network interface index.
//   - interface-name zone “eth0”: hasZone: true, isNumeric false
//   - it is not checked if interface number or name is an up interface
func Zone(addr netip.Addr) (zone string, znum int, hasZone, isNumeric bool) {

	// retrieve Zone
	zone = addr.Zone()

	// no zone case
	if hasZone = zone != ""; !hasZone {
		return // no zone return: hasZone false, all other zero-value
	}

	// interpret as numeric string
	var err error
	znum, err = strconv.Atoi(zone)
	isNumeric = err == nil

	return // zone return: zone is value, hasZone true, znum isNumeric valid
}

// Addr46 converts an IPv4-mapped IPv6 address 4in6 IPv6 addresses to
// IPv4 for consistent IPv4/IPv6
//   - addr: any [netip.Addr] address, including invalid
//   - addr46: [netip.Addr] where any IPv6 “::ffff:0:0/96” is IPv4
//   - — addr46 is unchanged addr if addr not IPv4-mapped IPv6 address
//   - —
//   - IPv6 has a special class of addresses representing an IPv4 address
//     “::ffff:0:0/96”
//   - IPv6 “::ffff:192.0.2.128” represents the IPv4 address “192.0.2.128”
//   - Addr46 converts such IPv6 addresses to IPv4
func Addr46(addr netip.Addr) (addr46 netip.Addr) {

	// [netip.Addr.Is4In6] returns true if addr is in “::ffff:0:0/96”
	if addr.Is4In6() {
		// netip does conversion using array values, ie. no allocation
		//	- [netip.Addr.As4] returns 4-byte array of IPv4 address
		//	- [netip.Addr.AddrFrom4] returns netip.Addr IPv4 from 4-byte array
		addr46 = netip.AddrFrom4(addr.As4())
		return // IPv4 value return
	}

	// othwerwise, no change
	addr46 = addr

	return // unchanged return
}

// NetworkPrefixBitCount returns network prefix size in bits
// when addr interpreted as network mask
//   - addr: valid IP address
//   - bits: the number of leading ones in addr
//   - IPv4: 0–32, IPv6: 0–128
//   - — addr invalid: -1
//   - — mask corrupt: addr not zero or more leading one-bits
//     followed by all zero-bits: -1
//   - size of the netmask is [netip.Addr.BitLen]
func NetworkPrefixBitCount(addr netip.Addr) (bits int, err error) {

	// addr invalid case
	if !addr.IsValid() {
		err = perrors.ErrorfPF("invalid address")
		bits = -1
		return // invalid address return: bits -1
	}

	// byte-slice length 4 for IPv4 or 16 for IPv6
	var byts = addr.AsSlice()

	// true if a zero-bit was found in mask
	var foundZeroBit bool

	// count bits that are 1 from the high order bit until all zero-bits are found
	for _, byt := range byts {
		switch byt {
		case 255:

			// byte of all one-bits
			if !foundZeroBit {
				// add bit-count
				bits += 8
				continue
			}
			// one-bits after first zero-bit: corrupt
			err = perrors.ErrorfPF("netmask: 255-byte after first zero-bit")

		case 0:
			// now scanning for zero-bits: OK
			foundZeroBit = true
			continue
		default:

			// byte of mixed zero and one bits
			if foundZeroBit {
				err = perrors.NewPF("netmask: non-zero byte after first zero-bits")
				break // one-bits after first zero-bit: corrupt
			}
			foundZeroBit = true

			// count possible leading one-bits
			for byt != 0 {
				if byt&128 != 0 {
					bits++
				}
				// shift left drops the one-bit
				byt <<= 1
			}
			// end of one-bits
			// all bits from now one must be zero-bits

			if byt == 0 {
				continue // continue scanning for zero-bits
			}
			// one-bits after first zero-bit: corrupt
			err = perrors.NewPF("netmask: one-bits after first zero-bit")
		}

		// corrupt case
		bits = -1
		return // corrupt netmask return: bits -1
	}

	return // scanned all bytes good return
}
