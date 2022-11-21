/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"math"
	"net"
	"strconv"

	"github.com/haraldrudell/parl/perrors"
)

// IfIndex is a dynamic reference to a network interface on Linux systems
type IfIndex uint32

func NewIfIndex(value int) (ifIndex IfIndex, err error) {
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
func (ifIndex IfIndex) Interface() (*net.Interface, error) {
	return net.InterfaceByIndex(int(ifIndex))
}

// Interface gets net.Interface for ifIndex
func (ifIndex IfIndex) InterfaceIndex() (interfaceIndex int) {
	return int(ifIndex)
}

// Zone gets net.IPAddr.Zone string for ifIndex
func (ifIndex IfIndex) Zone() (zone string, err error) {
	if ifIndex.IsValid() {
		if iface, err := ifIndex.Interface(); err == nil { // may fail if interface already deleted
			zone = iface.Name
		} else {
			zone = strconv.Itoa(int(ifIndex))
		}
	}
	return
}

func (ifIndex IfIndex) String() (s string) {
	return "#" + strconv.Itoa(int(ifIndex))
}
