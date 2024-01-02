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

func AddrPortToUDPAddr2(addrPort netip.AddrPort) (addr *net.UDPAddr) {
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

// AddrPortFromAddr converts tcp-protocol net.Addr to netip.AddrPort
func AddrPortFromAddr(addr net.Addr) (near netip.AddrPort, err error) {
	// Addr is interface { Network() String() }
	//	- runtime type is *net.TCPAddr struct { IP IP; Port int; Zone string }
	var a, ok = addr.(*net.TCPAddr)
	if !ok {
		err = perrors.ErrorfPF("listener.Addr runtime type not *net.TCPAddr: %T", addr)
		return
	}
	var b netip.Addr
	if b, ok = netip.AddrFromSlice(a.IP); !ok {
		err = perrors.ErrorfPF("listener.Addr bad length: %d", len(a.IP))
		return
	} else if a.Zone != "" {
		b = b.WithZone(a.Zone)
	}
	near = netip.AddrPortFrom(b, uint16(a.Port))

	return
}
