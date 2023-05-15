/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"errors"
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
func (ifIndex IfIndex) Interface() (netInterface *net.Interface, isErrNoSuchInterface bool, err error) {
	if netInterface, err = net.InterfaceByIndex(int(ifIndex)); err != nil {
		isErrNoSuchInterface = errors.Is(err, ErrNoSuchInterface)
		err = perrors.ErrorfPF("net.InterfaceByIndex %d %w", ifIndex, err)
	}
	return
}

// InterfaceAddrs gets Addresses for interface
//   - netInterface.Name is interface name "eth0"
//   - netInterface.Addr() returns assigned IP addresses
func (ifIndex IfIndex) InterfaceAddrs(useNameCache ...NameCacher) (name string, i4, i6 []netip.Prefix, err error) {
	var doCache = NoCache
	if len(useNameCache) > 0 {
		doCache = useNameCache[0]
	}

	var netInterface *net.Interface
	var isErrNoSuchInterface bool
	if netInterface, isErrNoSuchInterface, err = ifIndex.Interface(); err != nil {
		if isErrNoSuchInterface && doCache != NoCache {
			name, err = networkInterfaceNameCache.CachedName(ifIndex, doCache)
		}
		return // error or from cache
	}

	name = netInterface.Name
	i4, i6, err = InterfaceAddrs(netInterface)
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
		if iface, _, err = ifIndex.Interface(); err == nil { // may fail if interface already deleted
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
