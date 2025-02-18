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

func Interfaces() (interfaces []net.Interface, err error) {
	if interfaces, err = net.Interfaces(); err != nil {
		perrors.ErrorfPF("net.Interfaces %w", err)
	}
	return
}

// InterfaceAddrs gets Addresses for interface
//   - netInterface.Name is interface name "eth0"
//   - netInterface.Addr() returns assigned IP addresses
func InterfaceAddrs(netInterface *net.Interface) (i4, i6 []netip.Prefix, err error) {

	// get assigned IPv4 and IPv6 addresses
	var netAddrSlice []net.Addr
	if netAddrSlice, err = netInterface.Addrs(); perrors.IsPF(&err, "netInterface.Addrs %w", err) {
		return
	}
	// netAddr is interface with Network() String()
	//	- go1.20.3: type is *net.IPNet
	//	- Network is "ip+net"
	//	- String is "127.0.0.1/8" "fe80::1/64"
	var prefixes []netip.Prefix

	if prefixes, err = AddrSlicetoPrefix(netAddrSlice); err != nil {
		return
	}
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
//   - returns interface and prefix or error
//   - first uses Zone, then scans all interfaces for prefixes
//   - addr must be valid
func InterfaceFromAddr(addr netip.Addr) (netInterface *net.Interface, prefix netip.Prefix, isNoSuchInterface bool, err error) {

	// ensure there is IP
	if !addr.IsValid() {
		err = perrors.NewPF("addr cannot be invalid")
		return
	}

	// try zone
	zone, znum, hasZone, isNumeric := Zone(addr)
	if hasZone {
		if !isNumeric {
			if netInterface, err = net.InterfaceByName(zone); perrors.IsPF(&err, "net.InterfaceByName zone: %q %w", zone, err) {
				isNoSuchInterface = errors.Is(err, ErrNoSuchInterface)
				return
			}
		} else {
			if netInterface, err = net.InterfaceByIndex(znum); perrors.IsPF(&err, "net.InterfaceByName zone-numeric: %d %w", znum, err) {
				isNoSuchInterface = errors.Is(err, ErrNoSuchInterface)
				return
			}
		}
		if prefix, err = InterfacePrefix(netInterface, addr); err != nil {
			return
		}
	} else {

		// scan interfaces to find a network prefix
		var interfaces []net.Interface
		if interfaces, err = net.Interfaces(); perrors.IsPF(&err, "net.Interfaces %w", err) {
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
