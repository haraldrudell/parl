/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net/netip"
	"strconv"
	"strings"

	"github.com/haraldrudell/parl/perrors"
)

// Destination represents a selector for routing, ie. an IPv4 or IPv6 address with zone
// and prefix.
// go1.18 introduced [netip.Prefix] for this purpose
// see Linux: ip route add.
// [ root PREFIX ] [ match PREFIX ] [ exact PREFIX ] [ table TABLE_ID ] [ vrf NAME ]
// [ proto RTPROTO ] [ type TYPE ] [ scope SCOPE ]
// contains IP, Zone and Mask
type Destination struct {
	netip.Prefix
}

// NewDestination instantiates Destination.
// addr is IPv4 address or IPv6 address with Zone.
// prefix is number of bits actually used 0…32 for IPv4, 0…128 for IPv6
func NewDestination(prefix netip.Prefix) (d *Destination, err error) {
	if !prefix.IsValid() {
		err = perrors.ErrorfPF("invalid prefix: %#v", prefix)
		return
	}
	d0 := Destination{
		Prefix: prefix,
	}
	d = &d0
	return
}

// Key is a string suitable as a key in a map
func (d Destination) Key() (key netip.Prefix) {
	return d.Prefix
}

func (d Destination) String() (s string) {

	// abbreviate IPv4: 0.0.0.0/0 → 0/0 127.0.0.0/8 → 127/8
	if d.Prefix.Addr().Is4() {
		ipv4 := d.Prefix.Addr().String()
		for strings.HasSuffix(ipv4, ".0") {
			ipv4 = strings.TrimSuffix(ipv4, ".0")
		}
		return ipv4 + "/" + strconv.Itoa(d.Bits())
	}
	return d.Prefix.String()
}
