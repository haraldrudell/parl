/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/haraldrudell/parl"
)

const (
	UseInterfaceNameCache = true
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

// NewNextHopCounts returns NextHop with current IP address counts
//   - if input LinkAddr does not have interface name, interface name is added to output nextHop
//   - 6in4 are converted to IPv4
func NewNextHopCounts(gateway netip.Addr, linkAddr *LinkAddr, src netip.Addr,
	useNameCache ...NameCacher,
) (nextHop *NextHop, err error) {
	var doCache = NoCache
	if len(useNameCache) > 0 {
		doCache = useNameCache[0]
	}

	// tentative NextHop value based on input arguments
	var nh = NewNextHop(gateway, linkAddr, src)

	// interface from LinkAddr is required to get IP addresses assigned to the network interface
	var netInterface *net.Interface
	var interfaceName string
	var unknownInterface bool
	// obtain interface using index, name or mac
	if netInterface, unknownInterface, err = nh.LinkAddr.Interface(); err != nil {
		// for netlink packets of deleted interface, the index is already invalid
		//	- use the cache to obtain the interface name
		if unknownInterface && doCache != NoCache && linkAddr.IfIndex.IsValid() {
			if interfaceName, err = networkInterfaceNameCache.CachedName(linkAddr.IfIndex, doCache); err == nil {
				if nh.LinkAddr.Name == "" && interfaceName != "" {
					nh.LinkAddr.Name = interfaceName // update interface name if not already set
				}
			}
		}
		if err != nil {
			return // error in LinkAddr.Interface or CachedName
		}
	}
	if netInterface == nil {
		nextHop = nh
		return // Linkaddr interface did not exist
	}

	// update nh.Linkaddr from netInterface
	if !nh.LinkAddr.IfIndex.IsValid() {
		if nh.LinkAddr.IfIndex, err = NewIfIndexInt(netInterface.Index); err != nil {
			return
		}
	}
	if nh.LinkAddr.Name == "" {
		nh.LinkAddr.Name = netInterface.Name // update interface name if not already set
	}
	if len(nh.LinkAddr.HardwareAddr) == 0 {
		nh.LinkAddr.HardwareAddr = netInterface.HardwareAddr
	}

	// get IP address counts from interface
	var i4, i6 []netip.Prefix
	if i4, i6, err = InterfaceAddrs(netInterface); err != nil {
		return
	}
	nh.nIPv4 = len(i4)
	nh.nIPv6 = len(i6)
	// macOS lo0 has address:
	// for i := 0; i < len(i6); {
	// 	addr := i6[i].Addr()
	// 	if addr.Is4In6() {

	// 		i4 = append(i4, netip.PrefixFrom(addr.As4(), )
	// 		i6 = slices.Delete[](i6, i, i+1)
	// 		continue
	// 	}
	// 	i++
	// }
	nextHop = nh
	return
}

// NewNextHop2 assembles a route destination based on IfIndex
func NewNextHop2(index IfIndex, gateway netip.Addr, src netip.Addr) (next *NextHop, err error) {
	var linkAddr *LinkAddr
	if index.IsValid() {
		linkAddr = NewLinkAddr(index, "")
		if linkAddr, err = linkAddr.UpdateName(); err != nil {
			return
		}
	}
	return NewNextHop(gateway, linkAddr, src), err
}

func NewNextHop3(gateway netip.Addr, linkAddr *LinkAddr, src netip.Addr, nIPv4, nIPv6 int) (nextHop *NextHop) {
	return &NextHop{Gateway: gateway,
		LinkAddr: *linkAddr,
		Src:      src,
		nIPv4:    nIPv4,
		nIPv6:    nIPv6,
	}
}

// HasGateway determines if next hop uses a remote gateway
func (nh *NextHop) HasGateway() bool {
	return nh.Gateway.IsValid() && !nh.Gateway.IsUnspecified()
}

// HasSrc determines if next hop has src specified
func (nh *NextHop) HasSrc() bool {
	return nh.Src.IsValid() && !nh.Src.IsUnspecified()
}

func (n *NextHop) IsZeroValue() (isZeroValue bool) {
	return !n.Gateway.IsValid() &&
		n.LinkAddr.IsZeroValue() &&
		!n.Src.IsValid() &&
		n.nIPv4 == 0 &&
		n.nIPv6 == 0
}

// Name returns nextHop interface name
//   - name can be returned empty
//   - name of Linkaddr, then interface from index, name, mac
func (n *NextHop) Name(useNameCache ...NameCacher) (name string, err error) {
	var doCache = NoCache
	if len(useNameCache) > 0 {
		doCache = useNameCache[0]
	}

	// is name already present?
	if name = n.LinkAddr.Name; name != "" {
		return // nexthop had interface name available return
	}

	// interface from LinkAddr
	var netInterface *net.Interface
	var noSuchInterface bool
	if netInterface, noSuchInterface, err = n.LinkAddr.Interface(); err != nil {
		if noSuchInterface {
			err = nil
		} else {
			return // interface retrieval error return
		}
	}
	if netInterface != nil {
		name = netInterface.Name
		return // name from interface return
	}

	// interface from IP address
	var a = n.Gateway
	if !a.IsValid() {
		a = n.Src
	}
	if !a.IsValid() {
		return // no IP available return
	}
	if netInterface, _, _, err = InterfaceFromAddr(a); err != nil {
		return
	} else if netInterface != nil {
		name = netInterface.Name
		return
	}

	zone, znum, hasZone, isNumeric := Zone(a)
	if hasZone && !isNumeric {
		name = zone
		return // interface name from zone
	}

	// should cache be used?
	if doCache == NoCache {
		return
	}

	// get indexes that can be used with cache
	var ixs []IfIndex
	if n.LinkAddr.IfIndex.IsValid() {
		ixs = append(ixs, n.LinkAddr.IfIndex)
	}
	if isNumeric {
		var ifIndex IfIndex
		if ifIndex, err = NewIfIndexInt(znum); err != nil {
			return
		}
		if ifIndex.IsValid() && ifIndex != n.LinkAddr.IfIndex {
			ixs = append(ixs, ifIndex)
		}
	}

	// search cache
	for _, ifi := range ixs {
		if name, err = networkInterfaceNameCache.CachedName(ifi, doCache); err != nil {
			return
		} else if name != "" {
			return
		}
	}
	return
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

func (n *NextHop) Dump() (s string) {
	return parl.Sprintf("nextHop_gwIP_%s_%s_src_%s_4:%d_6:%d",
		n.Gateway.String(),
		n.LinkAddr.Dump(),
		n.Src,
		n.nIPv4, n.nIPv6,
	)
}

func (nextHop *NextHop) String() (s string) {

	// addr and hasNameZone
	var hasNameZone bool
	if nextHop.HasGateway() {
		s = nextHop.Gateway.String()
		gatewayAddr := nextHop.Gateway
		_, _, hasZone, isNumeric := Zone(gatewayAddr)
		hasNameZone = hasZone && !isNumeric
	}

	// interface name
	if !hasNameZone && !nextHop.LinkAddr.IsZeroValue() {
		if s != "" {
			s += "\x20"
		}
		s += nextHop.LinkAddr.OneString() // name or mac or if-index
	}

	// src 1.2.3.4
	if nextHop.Src.IsValid() &&
		((nextHop.Src.Is4() && nextHop.nIPv4 > 1) ||
			(nextHop.Src.Is6() && nextHop.nIPv6 > 1)) {
		s += " src " + nextHop.Src.String()
	}

	return
}
