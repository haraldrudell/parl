/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"fmt"
	"net"
	"net/netip"
	"strings"

	"github.com/haraldrudell/parl/perrors"
)

// NextHop describes a route target
type NextHop struct {
	/*
		if NextHop is an address on the local host or on a local subnet,
		Gateway is nil
		LinkAddr describes the local interface
		Src is the address on that local interface

		If Nexthop is remote, beyond any local subnet,
		Gateway is an IP on a local subnet
		LinkAddr describes the local interface for that subnet
		Src is the address on that local interface
	*/
	Gateway  netip.Addr
	LinkAddr            // LinkAddr is the hosts’s network interface where to send packets
	Src      netip.Addr // the source ip to use for originating packets on LinkAddr
	nIPv6    int        // number of assigned IPv6 addresses, 0 if unknown
	nIPv4    int        // number of assigned IPv6 addresses, 0 if unknown
}

// NewNextHop assembles a route destination
func NewNextHop(gateway netip.Addr, linkAddr *LinkAddr, src netip.Addr) (nextHop *NextHop) {
	next0 := NextHop{}
	if gateway.IsValid() {
		next0.Gateway = gateway
	}
	if linkAddr != nil {
		next0.LinkAddr = *linkAddr
	}
	if src.IsValid() {
		next0.Src = src
	}
	nextHop = &next0
	return
}

func NewNextHopCounts(gateway netip.Addr, linkAddr *LinkAddr, src netip.Addr) (nextHop *NextHop, err error) {

	// obatin interface by index
	var netInterface *net.Interface
	if linkAddr.IfIndex != 0 {
		if netInterface, err = net.InterfaceByIndex(int(linkAddr.IfIndex)); perrors.IsPF(&err, "net.InterfaceByIndex(%d) %w", linkAddr.IfIndex, err) {
			// on delete of interface, this ill fail
			// route ip+net: no such network interface
			// *net.OpError *errors.errorString
			// &net.OpError{Op:"route", Net:"ip+net", Source:net.Addr(nil), Addr:net.Addr(nil), Err:(*errors.errorString)(0x1400011a490)}
			err = nil // best effort
		}
	}
	if netInterface == nil {
		nextHop = NewNextHop(gateway, linkAddr, src)
		return // no interface return
	}

	// interface name is typically missing, populate it
	linkAddr2 := linkAddr
	if linkAddr.Name == "" {
		if linkAddr2, err = NewLinkAddr(linkAddr.IfIndex, netInterface.Name, linkAddr.HardwareAddr); err != nil {
			return
		}
	}

	nextHop0 := NewNextHop(gateway, linkAddr2, src)

	// we need to find how many IP addresses the interface has
	var netAddrSlice []net.Addr
	if netAddrSlice, err = netInterface.Addrs(); perrors.IsPF(&err, "netInterface.Addrs %w", err) {
		return
	}
	var ipv4 int
	var ipv6 int
	for _, a := range netAddrSlice {
		ipString := a.String()
		if index := strings.Index(ipString, "/"); index != -1 {
			ipString = ipString[:index]
		}
		var ip netip.Addr
		if ip, err = netip.ParseAddr(ipString); perrors.IsPF(&err, "netip.ParseAddr %w", err) {
			return
		}
		if ip.Is4() {
			ipv4++
		} else {
			ipv6++
		}
	}
	nextHop0.nIPv4 = ipv4
	nextHop0.nIPv6 = ipv6

	nextHop = nextHop0
	return
}

// NewNextHop2 assembles a route destination based on IfIndex
func NewNextHop2(index IfIndex, gateway netip.Addr, src netip.Addr) (next *NextHop, err error) {
	var linkAddr *LinkAddr
	if index.IsValid() {
		linkAddr, _ = NewLinkAddr(index, "", nil)
		if linkAddr, err = linkAddr.UpdateName(); err != nil {
			return
		}
	}
	return NewNextHop(gateway, linkAddr, src), err
}

// HasGateway determines if next hop uses a remote gateway
func (nh *NextHop) HasGateway() bool {
	return nh.Gateway.IsValid() && !nh.Gateway.IsUnspecified()
}

// HasSrc determines if next hop has src specified
func (nh *NextHop) HasSrc() bool {
	return nh.Src.IsValid() && !nh.Src.IsUnspecified()
}

// EmptyNextHop provides empty NextHop
func EmptyNextHop() *NextHop {
	return &NextHop{}
}

// Target describes the detination for this next hop
func (nh *NextHop) Target() (gateway netip.Addr, s string) {
	if nh == nil {
		return
	}
	s = nh.LinkAddr.OneString()
	if !nh.HasGateway() {
		return
	}
	gw := nh.Gateway
	gateway = gw
	if !gateway.IsValid() || gateway.IsUnspecified() {
		return
	}
	srcIP := nh.Src
	s1 := srcIP.String()
	index := nh.LinkAddr.IfIndex
	if index > 0 {
		iface, err := net.InterfaceByIndex(int(index))
		if err == nil {
			addrs, e2 := iface.Addrs()
			if e2 == nil {
				if srcIP.IsValid() && !srcIP.IsUnspecified() {
					if len(addrs) > 0 {
						s1 = addrs[0].String()
					}
				} else {
					for _, ipNet := range addrs {
						if ipNet, ok := ipNet.(*net.IPNet); ok {
							var netipAddr netip.Addr
							var ok bool
							if netipAddr, ok = netip.AddrFromSlice(ipNet.IP); !ok {
								continue
							}
							if netipAddr.Compare(srcIP) == 0 {
								s1 = ipNet.String()
								break
							}
						}
					}
				}
			}
		}
	}
	s = fmt.Sprintf("%s %s %s", gateway, s, s1)
	return
}

func (nextHop *NextHop) String() (s string) {
	if nextHop.HasGateway() {
		s = nextHop.Gateway.String() + "\x20"
	}
	s += nextHop.LinkAddr.OneString()
	if nextHop.Src.IsValid() &&
		((nextHop.Src.Is4() && nextHop.nIPv4 != 1) ||
			(nextHop.Src.Is6() && nextHop.nIPv6 != 1)) {
		s += " src " + nextHop.Src.String()
	}
	return
}
