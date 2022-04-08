/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"
	"strconv"
)

// IfIndex is a dynamic reference to a network interface on Linux systems
type IfIndex int

// Present determines if intreface index is set
func (ifIndex IfIndex) Present() bool {
	return ifIndex > 0
}

// Interface gets net.Interface for ifIndex
func (ifIndex IfIndex) Interface() (*net.Interface, error) {
	return net.InterfaceByIndex(int(ifIndex))
}

// Zone gets net.IPAddr.Zone string for ifIndex
func (ifIndex IfIndex) Zone() (zone string, err error) {
	if ifIndex.Present() {
		if iface, err := ifIndex.Interface(); err == nil { // may fail if interface already deleted
			zone = iface.Name
		} else {
			zone = strconv.Itoa(int(ifIndex))
		}
	}
	return
}
