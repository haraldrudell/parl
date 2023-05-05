/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import "net"

func DumpHardwareAddr(a net.HardwareAddr) (s string) {
	if len(a) == 0 {
		return ":"
	}
	return a.String()
}
