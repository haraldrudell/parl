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
func NewDestination(prefix netip.Prefix) (d *Destination) {
	return &Destination{Prefix: prefix}
}

func (d Destination) IsValid() (err error) {
	if !d.Prefix.IsValid() {
		err = perrors.ErrorfPF("invalid prefix: %#v", d.Prefix)
		return
	}
	return
}

// Key is a string suitable as a key in a map
func (d Destination) Key() (key netip.Prefix) {
	return d.Prefix
}

// IsDefaultRoute returns whether Destination is a default route:
//   - IPv4: 0/0 or 0.0.0.0/0
//   - IPv6: ::
func (d Destination) IsDefaultRoute() (isDefault bool) {
	var addr = d.Addr()
	if !addr.IsValid() || d.Bits() != 0 {
		return // address not valid or not /0
	}
	isDefault = addr.IsUnspecified()

	return
}

// "1.2.3.4/24" or "2000::/3"
// - abbreviate IPv4: "127.0.0.0/8" → "127/8"
// - IPv4 default route: "0.0.0.0/0" → "0/0"
//   - IPv6 default route stays "::/0"
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
