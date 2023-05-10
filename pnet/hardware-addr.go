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

func DumpHardwareAddr(a net.HardwareAddr) (s string) {
	if len(a) == 0 {
		return ":"
	}
	return a.String()
}

func HardwareAddrInterface(a net.HardwareAddr) (netInterface *net.Interface, isErrNoSuchInterface bool, err error) {
	if len(a) == 0 {
		err = perrors.NewPF("HarwareAddr cannot be empty")
		return
	}
	var interfaces []net.Interface
	if interfaces, err = net.Interfaces(); err != nil {
		return
	}
	for i := 0; i < len(interfaces); i++ {
		var iface = &interfaces[i]
		if slices.Equal(a, iface.HardwareAddr) {
			netInterface = iface
			return
		}
	}
	isErrNoSuchInterface = true
	err = perrors.ErrorfPF("No interface has mac: %s %w", a, ErrNoSuchInterface)
	return
}
