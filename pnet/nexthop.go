/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"fmt"
	"net"
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
	Gateway net.IPAddr
	LinkAddr
	Src net.IPAddr // the source ip to use on LinkAddr
}

// NewNextHop assembles a route destination
func NewNextHop(gateway *net.IPAddr, linkAddr *LinkAddr, src *net.IPAddr) *NextHop {
	next0 := NextHop{}
	if gateway != nil {
		next0.Gateway = *gateway
	}
	if linkAddr != nil {
		next0.LinkAddr = *linkAddr
	}
	if src != nil {
		next0.Src = *src
	}
	return &next0
}

// NewNextHop2 assembles a route destination based on IfIndex
func NewNextHop2(index IfIndex, gateway *net.IPAddr, src *net.IPAddr) (next *NextHop, err error) {
	var linkAddr *LinkAddr
	if index.Present() {
		linkAddr = NewLinkAddr(index)
		if err = linkAddr.UpdateName(); err != nil {
			return
		}
	}
	return NewNextHop(gateway, linkAddr, src), err
}

// HasGateway determines if next hop uses a remote gateway
func (nh *NextHop) HasGateway() bool {
	return IsNzIP(nh.Gateway.IP)
}

// HasSrc determines if next hop has src specified
func (nh *NextHop) HasSrc() bool {
	return IsNzIP(nh.Src.IP)
}

// EmptyNextHop provides empty NextHop
func EmptyNextHop() *NextHop {
	return &NextHop{}
}

// Target describes the detination for this next hop
func (nh *NextHop) Target() (gateway *net.IPAddr, s string) {
	if nh == nil {
		return
	}
	s = nh.LinkAddr.OneString()
	if !nh.HasGateway() {
		return
	}
	gw := nh.Gateway
	gateway = &gw
	if gateway == nil {
		return
	}
	srcIP := nh.Src.IP
	s1 := srcIP.String()
	index := nh.LinkAddr.Index
	if index > 0 {
		iface, err := net.InterfaceByIndex(int(index))
		if err == nil {
			addrs, e2 := iface.Addrs()
			if e2 == nil {
				if !IsNzIP(srcIP) {
					if len(addrs) > 0 {
						s1 = addrs[0].String()
					}
				} else {
					for _, ipNet := range addrs {
						if ipNet, ok := ipNet.(*net.IPNet); ok {
							if ipNet.IP.Equal(srcIP) {
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

func (nh NextHop) String() (s string) {
	if nh.HasGateway() {
		s = nh.Gateway.String() + "\x20"
	}
	s += fmt.Sprintf("%s %s", nh.LinkAddr.OneString(), &nh.Src)
	return
}
