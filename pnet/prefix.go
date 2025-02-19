/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net/netip"
)

// IPv4DefaultNetwork is the default route for IPv4: “0.0.0.0/0”
var IPv4DefaultNetwork = netip.MustParsePrefix("0.0.0.0/0")

// IPv6DefaultNetwork is the default route for IPv6: “::/0”
var IPv6DefaultNetwork = netip.MustParsePrefix("::/0")

// Prefix46 unmaps an IPv4-mapped IPv6 address into IPv4
//   - prefix: unchanged if IPv4, invalid or IPv6 other than “::ffff:0:0/96”
//   - — otherwise: “::ffff:0:0:1.2.3.4” → “1.2.3.4”
func Prefix46(prefix netip.Prefix) (prefix46 netip.Prefix) {

	// extract addr from prefix
	var addr = prefix.Addr()

	// Addr46 begins with Is4In6 so Is4In6 is faster
	if !addr.Is4In6() {
		prefix46 = prefix
		return
	}

	// convert to IPv4
	prefix46 = netip.PrefixFrom(Addr46(addr), prefix.Bits())

	return
}
