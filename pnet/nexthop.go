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
	"github.com/haraldrudell/parl/perrors"
)

const (
	UseInterfaceNameCache = true
)

// NextHop describes a route target
//   - NextHop describes how a packet is to be routed from this host
//     towards its final destination
//   - NextHop is typically obtained by looking up an IP address in
//     a routing table, ie. a collection of routing prefixes like “127/8” “::1”
//   - NexHop is either:
//   - — a local address for the host itself assigned to a local network interface
//   - — a subnet address for a host directly reachable via a local network interface
//   - — a gateway address on a subnet directly reachable via a local network interface
type NextHop struct {
	// Gateway is set if a gateway is used to route beyond local subnets
	//	- if NextHop is an address on the local host or on a local subnet:
	//	- — Gateway is nil
	//	- — LinkAddr describes the local interface
	//	- — Src is the address on that local interface
	//	- If Nexthop is remote, beyond any local subnet:
	//	- — Gateway is an IP on a local subnet
	//	- — LinkAddr describes the local interface for that subnet
	//	- — Src is the address on that local interface
	Gateway netip.Addr
	// LinkAddr is the hosts’s network interface where to send packets
	LinkAddr
	// Src is the source ip to use for originating packets on LinkAddr
	Src netip.Addr
	// nIPv6 is the number of assigned IPv6 addresses, 0 if unknown
	nIPv6 int
	// nIPv4 is the number of assigned IPv6 addresses, 0 if unknown
	nIPv4 int
}

// NewNextHop assembles a route destination
//   - gateway: optional gateway IP
//   - linkAddr: optional local network interface
//   - src: optional source IP to use on linkAddr
func NewNextHop(gateway netip.Addr, linkAddr *LinkAddr, src netip.Addr) (nextHop *NextHop) {
	var next0 = NextHop{}
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
//   - gateway: optional gateway IP
//   - linkAddr: optional local network interface
//   - src: optional source IP to use on linkAddr
//   - useNameCache missing: name cache is not used
//   - useNameCache [pnet.Update]: name cache is used and first updated
//   - useNameCache [pnet.NoUpdate]: name cache is used without update
//     -
//   - the name cache contains local network interface names for interfaces that are no longer up
//   - if input LinkAddr does not have interface name, interface name is added to output nextHop
//   - 6in4 are converted to IPv4
func NewNextHopCounts(
	gateway netip.Addr,
	linkAddr *LinkAddr,
	src netip.Addr,
	useNameCache ...NameCacher,
) (nextHop *NextHop, err error) {

	// default: no cache used
	var cacheParameter = NoCache
	if len(useNameCache) > 0 {
		cacheParameter = useNameCache[0]
	}

	// tentative NextHop value based on input arguments
	var nextHop0 = NewNextHop(gateway, linkAddr, src)

	// obtain network interface from provided LinkAddr
	//	- searched using LinkAddr index, name then mac
	//	- result contains interface index, name and IP address assignments
	var netInterface *net.Interface
	// network interface name obtained via LinkAddr index and cache
	var interfaceName string
	if netInterface, interfaceName, err = getInterface(&nextHop0.LinkAddr, cacheParameter); err != nil {
		return // error in LinkAddr.Interface or CachedName
	} else if netInterface == nil {
		if nextHop0.LinkAddr.Name == "" && interfaceName != "" {
			nextHop0.LinkAddr.Name = interfaceName // update interface name if not already set
		}
		nextHop = nextHop0
		return // Linkaddr interface did not exist
	}
	// netInterface is present

	// update nh.Linkaddr from netInterface
	if !nextHop0.LinkAddr.IfIndex.IsValid() {
		if nextHop0.LinkAddr.IfIndex, err = NewIfIndexInt(netInterface.Index); err != nil {
			return
		}
	}
	if nextHop0.LinkAddr.Name == "" {
		nextHop0.LinkAddr.Name = netInterface.Name // update interface name if not already set
	}
	if len(nextHop0.LinkAddr.HardwareAddr) == 0 {
		nextHop0.LinkAddr.HardwareAddr = netInterface.HardwareAddr
	}

	// get IP address counts from interface
	var i4, i6 []netip.Prefix
	if i4, i6, err = InterfaceAddrs(netInterface); err != nil {
		return
	}
	nextHop0.nIPv4 = len(i4)
	nextHop0.nIPv6 = len(i6)
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
	nextHop = nextHop0
	return
}

// getInterface gets netInterface or interfaceName for linkAddr
//   - searched using LinkAddr index, name then mac
//   - if interface not found but index valid and cacheParameter not NoCache:
//   - find interface name using cache
func getInterface(linkAddr *LinkAddr, cacheParameter NameCacher) (netInterface *net.Interface, interfaceName string, err error) {

	// obtain interface using index, name or mac
	var unknownInterface bool
	if netInterface, unknownInterface, err = linkAddr.Interface(); err == nil {
		return // interface obtained
	} else if !unknownInterface || cacheParameter == NoCache || !linkAddr.IfIndex.IsValid() {
		return // unfixable error return
	}
	// unknown interface but ifindex is valid and use cache

	// check the cache of interfaces that were once up
	// for netlink packets of deleted interface, the index is already invalid
	//	- use the cache mapping LinkAddr index to obtain the interface name
	switch cacheParameter {
	case Update:
		interfaceName, err = networkInterfaceNameCache.CachedName(linkAddr.IfIndex)
	case NoUpdate:
		interfaceName = networkInterfaceNameCache.CachedNameNoUpdate(linkAddr.IfIndex)
		err = nil
	}

	return
}

// NewNextHop2 assembles a route destination based on IfIndex
//   - index: optional interface index
//   - gateway: optional gateway to use for nextHop
//   - src: optional source address to use for nextHop
//   - err: index is valid but the network interface could not be retrieved
//   - — typically because it was deleted: adapter removed or VPN link down
func NewNextHop2(index IfIndex, gateway netip.Addr, src netip.Addr) (next *NextHop, err error) {

	// local network interface for nextHop
	var linkAddr *LinkAddr
	if index.IsValid() {
		linkAddr = NewLinkAddr(index, "")
		if linkAddr, err = linkAddr.UpdateName(); err != nil {
			return // retrieve network interface failed error return
		}
	}
	next = NewNextHop(gateway, linkAddr, src)

	return
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
func (n *NextHop) HasGateway() bool {
	return n.Gateway.IsValid() && !n.Gateway.IsUnspecified()
}

// HasSrc determines if next hop has src specified
func (n *NextHop) HasSrc() bool {
	return n.Src.IsValid() && !n.Src.IsUnspecified()
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

	// cache parameter default NoCache
	var doCache = NoCache
	if len(useNameCache) > 0 {
		doCache = useNameCache[0]
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
		switch doCache {
		case NoUpdate:
			name = networkInterfaceNameCache.CachedNameNoUpdate(ifi)
		case Update:
			if name, err = networkInterfaceNameCache.CachedName(ifi); err != nil {
				return
			}
		}
		if name != "" {
			return
		}
	}
	return
}

// Target describes the destination for this next hop
//   - gateway is invalid for local network targets or gateway unspecified address "0.0.0.0"
//   - s is:
//   - empty string for nil NextHop
//   - network interface name mac index or 0 for gateway missing, invalid or unspecified
//   - otherwise gateway ip, interface description and source IP or source cidr
func (n *NextHop) Target() (gateway netip.Addr, s string) {

	// ensure NextHop present
	if n == nil {
		return // no NextHop available: invalid and empty string
	}

	// network interface description:
	// “en5” or “aa:bb…” or “#3” or “0” but never empty string
	s = n.LinkAddr.OneString()
	if !n.HasGateway() {
		return // target is on local network, only the network interface describes it, gateway invalid
	}
	// gateway is valid gateway IP from NextHop field
	gateway = n.Gateway
	// is4 indicates that an IPv4 address is sought
	var is4 = Addr46(gateway).Is4()

	// try to obtain source IP and prefix

	// srcIP is possible source IP specified in NextHop
	var srcIP netip.Addr
	// hasSrcIP indicates that srcIP is usable
	var hasSrcIP = n.Src.IsValid() && !n.Src.IsUnspecified()
	if hasSrcIP {
		srcIP = Addr46(n.Src)
	}
	var cidr netip.Prefix
	var e error
	cidr, e = n.ifCidr(srcIP, is4)
	_ = e

	// sourceIPString is source IP “1.2.3.4” or cidr “1.2.3.4/24”
	var sourceIPString string
	if cidr.IsValid() {
		sourceIPString = "\x20" + cidr.String()
	} else if hasSrcIP {
		sourceIPString = "\x20" + srcIP.String()
	}

	// “192.168.1.12 en5 192.168.1.21/24”
	s = fmt.Sprintf("%s %s%s", gateway, s, sourceIPString)

	return
}

// targets tries to obtain cidr from network interface IP assignment
func (n *NextHop) ifCidr(srcIP netip.Addr, is4 bool) (cidr netip.Prefix, err error) {

	// network interface index
	var index = n.LinkAddr.IfIndex
	if index == 0 {
		return // failed to obtain network interface index
	}
	var iface *net.Interface
	if iface, err = net.InterfaceByIndex(int(index)); perrors.IsPF(&err, "InterfaceByIndex %w", err) {
		return // failed to obtain interface
	}
	// addrs is a list of IP addresses assigned to the network interface
	var addrs []net.Addr
	if addrs, err = iface.Addrs(); perrors.IsPF(&err, "iface.Addrs %w", err) {
		return // failed to obtain interface address
	}
	if len(addrs) == 0 {
		return // failed to obtain any assigned addresses
	}
	// prefixes are list of cidrs assigned top network interface
	var prefixes []netip.Prefix
	if prefixes, err = AddrSlicetoPrefix(addrs, Do46Yes); err != nil {
		return // addr slice failed to convert
	}
	if !srcIP.IsValid() {
		// if no srcIP, pick the first address of the correct family
		for _, prefix := range prefixes {
			var addr = Addr46(prefix.Addr())
			if addr.Is4() == is4 {
				cidr = prefix
				return // found a cidr
			}
		}
		var family string
		if is4 {
			family = "IPv4"
		} else {
			family = "IPv6"
		}
		err = perrors.ErrorfPF("interface %s has no %d addresses", iface.Name, family)
		return
	}
	// find the prefix that contains scrIP, it should exist
	for _, prefix := range prefixes {
		if prefix.Contains(srcIP) {
			cidr = prefix
			return
		}
	}
	err = perrors.ErrorfPF("interface %s is not assign a cidr for %s", iface.Name, srcIP)
	return
}

// Dump displays all fields of NextHop for troubleshooting
func (n *NextHop) Dump() (s string) {
	return parl.Sprintf("nextHop_gwIP_%s_%s_src_%s_4:%d_6:%d",
		n.Gateway.String(),
		n.LinkAddr.Dump(),
		n.Src,
		n.nIPv4, n.nIPv6,
	)
}

func (n *NextHop) String() (s string) {

	// addr and hasNameZone
	var hasNameZone bool
	if n.HasGateway() {
		s = n.Gateway.String()
		gatewayAddr := n.Gateway
		_, _, hasZone, isNumeric := Zone(gatewayAddr)
		hasNameZone = hasZone && !isNumeric
	}

	// interface name
	if !hasNameZone && !n.LinkAddr.IsZeroValue() {
		if s != "" {
			s += "\x20"
		}
		s += n.LinkAddr.OneString() // name or mac or if-index
	}

	// src 1.2.3.4
	if n.Src.IsValid() &&
		((n.Src.Is4() && n.nIPv4 > 1) ||
			(n.Src.Is6() && n.nIPv6 > 1)) {
		s += " src " + n.Src.String()
	}

	return
}
