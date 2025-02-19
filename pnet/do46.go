/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import "net/netip"

// the [netip.Addr] method extracting IPv4 from IPv6
var _ = netip.Addr.Is4In6

const (
	// return an IPv4 address stored in IPv6 as the IPv4
	//   - IPv4-mapped addresses written like “::ffff:1.2.3.4”
	Do46Yes = true
	// return an IPv4 address stored in IPv6 as the IPv4
	//   - IPv4-mapped addresses written like “::ffff:1.2.3.4”
	Do46No = false
)

// Do46 describes whether IPv4 in IPv6 translation should be done [Do46Yes]
//   - [Do46No]
//   - an IPv4 address can be stored in an IPv6 address in network [::ffff:0:0/96]
//   - IPv4-mapped addresses written like “::ffff:1.2.3.4”
type Do46 bool
