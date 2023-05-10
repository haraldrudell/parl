/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import "net/netip"

var IPv4VpnPrefix = netip.MustParsePrefix("0.0.0.0/1")
var IPv6VpnPrefix = netip.MustParsePrefix("::/3")
