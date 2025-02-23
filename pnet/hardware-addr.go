/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"

	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/slices"
)

// DumpHardwareAddr returns printable non-empty string for [net.HardwareAddr]
//   - [net.HardwareAddr.String] that never returns empty string
//   - nil or zero-length hardware address: “:”
//   - otherwise “00:11:22:33:44:55”
//   - for troubleshooting purposes
func DumpHardwareAddr(a net.HardwareAddr) (s string) {
	if len(a) == 0 {
		return ":"
	}
	return a.String()
}

// HardwareAddrInterface returns the network interface that has
// provided hardware address
//   - a: non-empty hardware address “00:11:22:33:44:55”
//   - netInterface: the corresponding [*net.Interface]
//   - isErrNoSuchInterface true: a network interface with address a is not up; err is non-nil
//   - err: a is empty, route.FetchRIB route.ParseRIB error,
//     no interface up matching a
func HardwareAddrInterface(a net.HardwareAddr) (netInterface *net.Interface, isErrNoSuchInterface bool, err error) {

	// empty hardware address is error
	if len(a) == 0 {
		err = perrors.NewPF("HardwareAddr cannot be empty")
		return
	}

	// get list of up system interfaces
	var interfaces []net.Interface
	if interfaces, err = net.Interfaces(); err != nil {
		return
	}

	// find any interface with matching hardware address
	for i := 0; i < len(interfaces); i++ {
		var iface = &interfaces[i]
		if slices.Equal(a, iface.HardwareAddr) {
			netInterface = iface
			return // foound interface return: netInterface non-nil
		}
	}

	// no such interface
	isErrNoSuchInterface = true
	err = perrors.ErrorfPF("No interface has mac: %s %w", a, ErrNoSuchInterface)

	return
}
