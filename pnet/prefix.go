/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net/netip"
)

var IPv4DefaultNetwork = netip.MustParsePrefix("0.0.0.0/0")
var IPv6DefaultNetwork = netip.MustParsePrefix("::/0")

func Prefix46(prefix netip.Prefix) (prefix46 netip.Prefix) {
	var addr = prefix.Addr()
	if !addr.Is4In6() {
		prefix46 = prefix
	} else {
		prefix46 = netip.PrefixFrom(Addr46(addr), prefix.Bits())
	}
	return
}
