/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"math"
	"net"
	"net/netip"
	"strconv"

	"github.com/haraldrudell/parl/perrors"
)

// IfIndex is a dynamic reference to a network interface on Linux systems
type IfIndex uint32

func NewIfIndex(index uint32) (ifIndex IfIndex) {
	return IfIndex(index)
}

func NewIfIndexInt(value int) (ifIndex IfIndex, err error) {
	if value < 0 || value > math.MaxUint32 {
		err = perrors.ErrorfPF("Value not uint32: %d", value)
		return
	}
	ifIndex = IfIndex(value)
	return
}

// IsValid determines if interface index value is set, ie. > 0
func (ifIndex IfIndex) IsValid() (isValid bool) {
	return ifIndex > 0
}

// Interface gets net.Interface for ifIndex
//   - netInterface.Name is interface name "eth0"
//   - netInterface.Addr() returns assigned IP addresses
func (ifIndex IfIndex) Interface() (netInterface *net.Interface, err error) {
	netInterface, err = net.InterfaceByIndex(int(ifIndex))
	perrors.IsPF(&err, "net.InterfaceByIndex %d %w", ifIndex, err)
	return
}

// InterfaceAddrs gets Addresses for interface
//   - netInterface.Name is interface name "eth0"
//   - netInterface.Addr() returns assigned IP addresses
func (ifIndex IfIndex) InterfaceAddrs() (name string, i4, i6 []netip.Prefix, err error) {
	var netInterface *net.Interface
	if netInterface, err = ifIndex.Interface(); err != nil {
		return
	}
	name = netInterface.Name

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

// Interface gets net.Interface for ifIndex
//   - InterfaceIndex is unique for promting methods
func (ifIndex IfIndex) InterfaceIndex() (interfaceIndex int) {
	return int(ifIndex)
}

// Zone gets net.IPAddr.Zone string for ifIndex
//   - if an interfa ce name can be ontained, that is the zone
//   - otherwise a numeric zone is used
//   - if ifIndex is invalid, empty string
func (ifIndex IfIndex) Zone() (zone string, isNumeric bool, err error) {
	if ifIndex.IsValid() {
		var iface *net.Interface
		if iface, err = ifIndex.Interface(); err == nil { // may fail if interface already deleted
			zone = iface.Name
		} else {
			zone = strconv.Itoa(int(ifIndex))
			isNumeric = true
		}
	}
	return
}

// "#13"
func (ifIndex IfIndex) String() (s string) {
	return "#" + strconv.Itoa(int(ifIndex))
}
