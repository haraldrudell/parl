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

var _ net.Addr
var _ netip.Addr
var _ net.IPAddr

// AddrToIPAddr returns the [net.Addr] for an [netip.Addr]
//   - [net.Addr] is legacy interface for [net.Dial] using strings
//   - — [net.Addr.Network] is string like “tcp”
//   - — [net.Addr.String] is socket address literal like “1.2.3.4:80”
//   - [netip.Addr] is binary value-literals based on integers
//   - [net.IPAddr] is a legacy implementation of [net.Addr] for
//     IPv4 or IPv6 addresses
//   - — [net.IPAddr.Network] returns “ip”
//   - — [net.IPAddr.String] returns “1.2.3.4” or “fe80::%eth0”
func AddrToIPAddr(addr netip.Addr) (addrInterface net.Addr) {
	if !addr.IsValid() {
		panic(perrors.NewPF("invalid netip.Addr"))
	}
	return &net.IPAddr{IP: addr.AsSlice(), Zone: addr.Zone()}
}

// AddrPortToUDPAddr: Network() "udp"
func AddrPortToUDPAddr(addrPort netip.AddrPort) (addrInterface net.Addr) {
	IP, port, zone := SplitAddrPort(addrPort)
	return &net.UDPAddr{IP: IP, Port: port, Zone: zone}
}

// IPAddr returns IPAddr from IP and IfIndex to IPAddr
func IPAddr(IP net.IP, index IfIndex, zone string) (ipa *net.IPAddr, err error) {
	ipa = &net.IPAddr{IP: IP}
	if IsIPv6(IP) {
		if zone != "" {
			ipa.Zone = zone
		} else {
			ipa.Zone, _, err = index.Zone()
		}
	}
	return
}

func AddrPortToUDPAddr2(addrPort netip.AddrPort) (addr *net.UDPAddr) {
	IP, port, zone := SplitAddrPort(addrPort)
	return &net.UDPAddr{IP: IP, Port: port, Zone: zone}
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

// AddrPortToTCPAddr: Network() "tcp"
func AddrPortToTCPAddr(addrPort netip.AddrPort) (addrInterface net.Addr) {
	IP, port, zone := SplitAddrPort(addrPort)
	return &net.TCPAddr{IP: IP, Port: port, Zone: zone}
}
