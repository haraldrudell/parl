/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"
	"net/netip"

	"github.com/haraldrudell/parl/perrors"
)

// AddrPortFromAddr converts tcp-protocol string-based [net.Addr]
// to binary [netip.AddrPort]
//   - addr should be [net.Addr] for tcp-network address
//     implemented by [net.TCPAddr]
//   - [netip.AddrPort] is a binary-coded socket address
//   - [net.Addr] is legacy for [net.Dial] using strings for Network and
//     socket address
//   - [net.TCPAddr] returns
//   - — Network “tcp”
//   - — String like “[fe80::%eth0]:80”
func AddrPortFromAddr(addr net.Addr) (near netip.AddrPort, err error) {
	// Addr is interface { Network() String() }
	//	- runtime type is *net.TCPAddr struct { IP IP; Port int; Zone string }
	var a, ok = addr.(*net.TCPAddr)
	if !ok {
		err = perrors.ErrorfPF("listener.Addr runtime type not *net.TCPAddr: %T", addr)
		return
	}
	var b netip.Addr
	if b, ok = netip.AddrFromSlice(a.IP); !ok {
		err = perrors.ErrorfPF("listener.Addr bad length: %d", len(a.IP))
		return
	} else if a.Zone != "" {
		b = b.WithZone(a.Zone)
	}
	near = netip.AddrPortFrom(b, uint16(a.Port))

	return
}
