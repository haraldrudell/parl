/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"errors"
	"net"
	"net/netip"

	"github.com/haraldrudell/parl/perrors"
)

// Interfaces returns a list of the system’s network interfaces
//   - err: error in route.FetchRIB or route.ParseRIB
//   - [net.Interfaces] with better error
func Interfaces() (interfaces []net.Interface, err error) {

	// error in route.FetchRIB or route.ParseRIB
	if interfaces, err = net.Interfaces(); err != nil {
		err = perrors.ErrorfPF("net.Interfaces %w", err)
	}

	return
}

// hook for testing
var netInterfaceAddrsHook func() (addrs []net.Addr, err error)

// InterfaceAddrs returned assigned network prefixes for an interface
//   - [net.Interface.Addrs] returning non-legacy [netip.Prefix] sorted by IPv4/IPv6
//   - netInterface: received from [net.InterfaceByIndex] [net.InterfaceByName] [pnet.Interfaces]
//   - i4, i6: any network prefixes assigned to netInterface for IPv4 and IPv6
//   - — IPv4-mapped IPv6 address is translated to IPv4
//   - err: netInterface nil, error in route.FetchRIB or route.ParseRIB
//   - err: corrupt adress data
//   - —
//   - delegates to [net.Interface.Addrs]
//   - macOS IPv4 may be IPv4-mapped IPv6 address:
//     128-bit “::ffff:127.0.0.1/8” with 32-bit netmask [255, 0, 0, 0]
func InterfaceAddrs(netInterface *net.Interface) (i4, i6 []netip.Prefix, err error) {

	// get assigned IPv4 and IPv6 network prefixes
	//	- [net.Addr] implemented by [*net.IPNet]
	var netAddrSlice []net.Addr
	if n := netInterfaceAddrsHook; n == nil {
		// err: netInterface nil, error in route.FetchRIB or route.ParseRIB
		netAddrSlice, err = netInterface.Addrs()
	} else {
		netAddrSlice, err = n()
	}
	if err != nil {
		err = perrors.ErrorfPF("netInterface.Addrs %w", err)
		return // [net.Interface.Addrs] error return
	}

	// netAddr is interface with Network() String()
	//	- go1.20.3: implementing type is [*net.IPNet]
	//	- [net.IPNet.Network] is “ip+net”
	//	- [net.IPNet.String] is “127.0.0.1/8” “fe80::1/64”
	var prefixes []netip.Prefix

	// convert []net.Addr to []netip.Prefix
	if prefixes, err = AddrSlicetoPrefix(netAddrSlice); err != nil {
		return // corrupt address data error return
	}

	// separate prefixes into IPv4 and IPv6
	//	- IPv4-mapped IPv6 address is kept IPv6
	for _, prefix := range prefixes {
		if prefix.Addr().Is4() {
			i4 = append(i4, prefix)
		} else {
			i6 = append(i6, prefix)
		}
	}

	return
}

// InterfaceFromAddr finds network interface and prefix for addr
//   - addr: an IP address
//   - — if addr has zone with network interface name or
//     numeric network interface index, this is used
//   - netInterface: the network interface, if any, assigned addr
//   - isNoSuchInterface: a network interface name or index doe snot exist
//   - err:
//   - — addr is not valid
//   - — addr.Zone network interface name does not exist
//   - — addr.Zone numeric network interface index does not exist
//   - returns interface and prefix or error
//   - first uses Zone, then scans all interfaces for prefixes
//   - addr must be valid
func InterfaceFromAddr(addr netip.Addr) (netInterface *net.Interface, prefix netip.Prefix, isNoSuchInterface bool, err error) {

	// ensure there is IP
	if !addr.IsValid() {
		err = perrors.NewPF("addr cannot be invalid")
		return // addr invalid error return
	}

	// try zone
	var zone, znum, hasZone, isNumeric = Zone(addr)
	if hasZone {
		if !isNumeric {
			if netInterface, err = net.InterfaceByName(zone); perrors.IsPF(&err, "net.InterfaceByName zone: %q %w", zone, err) {
				isNoSuchInterface = errors.Is(err, ErrNoSuchInterface)
				return // bad zone name error return
			}
		} else {
			if netInterface, err = net.InterfaceByIndex(znum); perrors.IsPF(&err, "net.InterfaceByName zone-numeric: %d %w", znum, err) {
				isNoSuchInterface = errors.Is(err, ErrNoSuchInterface)
				return // bad zone index error return
			}
		}

		// find if addr is assigned to the specified network interface
		if prefix, err = InterfacePrefix(netInterface, addr); err != nil {
			return
		}
	} else {

		// scan interfaces to find a network prefix
		var interfaces []net.Interface
		if interfaces, err = Interfaces(); err != nil {
			return
		}
		for i := 0; i < len(interfaces); i++ {
			var ifp = &interfaces[i]
			if prefix, err = InterfacePrefix(ifp, addr); err != nil {
				return
			} else if prefix.IsValid() {
				netInterface = ifp
				break
			}
		}
	}

	if isNoSuchInterface = netInterface == nil; isNoSuchInterface {
		err = perrors.ErrorfPF("no network interface has address: %s", addr)
		return
	}

	if !prefix.IsValid() {
		err = perrors.ErrorfPF("network interface %s does not have address: %s", netInterface.Name, addr)
		return
	}

	return
}

// InterfacePrefix returns the network prefix assigned to netInterface that contains addr
//   - if addr is not part of any prefix, returned prefix is invalid
func InterfacePrefix(netInterface *net.Interface, addr netip.Addr) (prefix netip.Prefix, err error) {
	var i4, i6 []netip.Prefix
	if i4, i6, err = InterfaceAddrs(netInterface); err != nil {
		return
	}
	if addr.Is4() {
		for _, i4p := range i4 {
			if i4p.Contains(addr) {
				prefix = i4p
				return // interface name by finding assigned IP
			}
		}
		var a6 = netip.AddrFrom16(addr.As16())
		for _, i6p := range i6 {
			if i6p.Contains(a6) {
				prefix = i6p
				return // interface name by finding assigned IP
			}
		}
	} else {
		for _, i6p := range i6 {
			if i6p.Contains(addr) {
				prefix = i6p
				return // interface name by finding assigned IP
			}
		}
	}
	return
}
