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
	var netAddr net.Addr
	for _, netAddr = range netAddrSlice {
		var netIPNet *net.IPNet
		var ok bool
		if netIPNet, ok = netAddr.(*net.IPNet); !ok {
			err = perrors.ErrorfPF("type assertion failed actual: %T expected: %T", netAddr, netIPNet)
			return
		}
		var prefix netip.Prefix
		if prefix, err = IPNetToPrefix(netIPNet); err != nil {
			return
		}
		if prefix.Addr().Is4() {
			i4 = append(i4, prefix)
		} else {
			i6 = append(i6, prefix)
		}
	}

	return
}
