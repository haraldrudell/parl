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

// NewIfIndex returns the index of a local network interface
func NewIfIndex(index uint32) (ifIndex IfIndex) { return IfIndex(index) }

// NewIfIndexInt returns the index of a local network interface
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

// Interface gets [net.Interface] for ifIndex
//   - netInterface.Name is interface name "eth0"
//   - netInterface.Addr() returns assigned IP addresses
//   - isErrNoSuchInterface: error is that interface number ifIndex does not exist
//   - — typically because it was deleted: adapter removed or VPN link down
//   - err: [net.InterfaceByIndex]
func (ifIndex IfIndex) Interface() (netInterface *net.Interface, isErrNoSuchInterface bool, err error) {
	if netInterface, err = net.InterfaceByIndex(int(ifIndex)); err != nil {
		isErrNoSuchInterface = errors.Is(err, ErrNoSuchInterface)
		err = perrors.ErrorfPF("net.InterfaceByIndex %d %w", ifIndex, err)
	}
	return
}

// InterfaceAddrs gets Addresses for interface augmented with cache
//   - useNameCache missing: cache of previously up neetwork interfaces is not used
//   - useNameCache [pnet.Update]: cache used after update
//   - useNameCache [pnet.NoUpdate]: cache used without update
//   - netInterface.Name is interface name “eth0”
//   - netInterface.Addr() returns assigned IP addresses
func (ifIndex IfIndex) InterfaceAddrs(useNameCache ...NameCacher) (name string, i4, i6 []netip.Prefix, err error) {

	// whether to use cache
	var doCache = NoCache
	if len(useNameCache) > 0 {
		doCache = useNameCache[0]
	}

	var netInterface *net.Interface
	var isErrNoSuchInterface bool
	if netInterface, isErrNoSuchInterface, err = ifIndex.Interface(); err != nil {
		if isErrNoSuchInterface {
			switch doCache {
			case NoCache:
			case Update:
				name, err = networkInterfaceNameCache.CachedName(ifIndex)
			case NoUpdate:
				name = networkInterfaceNameCache.CachedNameNoUpdate(ifIndex)
				err = nil
			}
		}
		return // error or from cache
	}

	name = netInterface.Name
	i4, i6, err = InterfaceAddrs(netInterface)

	return
}

// Interface gets net.Interface for ifIndex
//   - InterfaceIndex is unique for promting methods
func (ifIndex IfIndex) InterfaceIndex() (interfaceIndex int) { return int(ifIndex) }

// Zone gets net.IPAddr.Zone string for ifIndex
//   - if a network interface name can be obtained, that is the zone
//   - if network interface name could be obtained:
//   - — zone non-empty, isNumeric false, err nil
//   - if using numeric string:
//   - — zone non-empty, isNumeric true, err non-nil
//   - ifIndex invalid: zero-values
func (ifIndex IfIndex) Zone() (zone string, isNumeric bool, err error) {

	// if no index available: zero-values
	if !ifIndex.IsValid() {
		return
	}

	// get network interface name or numeric value from index
	var iface *net.Interface
	// may fail if interface already deleted
	if iface, _, err = ifIndex.Interface(); err == nil {
		zone = iface.Name
		if zone != "" {
			return
		}
		err = perrors.ErrorfPF("empty name for #%d", ifIndex.InterfaceIndex())
	}
	// err is non-nil

	// use numeric string
	zone = strconv.Itoa(ifIndex.InterfaceIndex())
	isNumeric = true

	return
}

// “#13”
func (ifIndex IfIndex) String() (s string) {
	return "#" + strconv.Itoa(int(ifIndex))
}
