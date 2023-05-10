/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"
	"net/netip"

	"github.com/haraldrudell/parl/perrors"
)

// AddrPortToUDPAddr: Network() "udp"
func AddrPortToUDPAddr(addrPort netip.AddrPort) (addrInterface net.Addr) {
	IP, port, zone := SplitAddrPort(addrPort)
	return &net.UDPAddr{IP: IP, Port: port, Zone: zone}
}

// AddrPortToTCPAddr: Network() "tcp"
func AddrPortToTCPAddr(addrPort netip.AddrPort) (addrInterface net.Addr) {
	IP, port, zone := SplitAddrPort(addrPort)
	return &net.TCPAddr{IP: IP, Port: port, Zone: zone}
}

func SplitAddrPort(addrPort netip.AddrPort) (IP net.IP, port int, zone string) {
	if !addrPort.IsValid() {
		panic(perrors.NewPF("invalid netip.AddrPort"))
	}
	IP = addrPort.Addr().AsSlice()
	port = int(addrPort.Port())
	zone = addrPort.Addr().Zone()
	return
}
