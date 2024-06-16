/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"fmt"
	"net/netip"
)

// Address holds an address for IP, TCP, UDP, domain-socket or derived protocols
//   - used for creating certificates
//   - created by [NewAddress]
//   - also:
//   - — [AddrPort]
//   - — [SocketAddress]
//   - — [Network]
type Address interface {
	// address as [netip.Addr]
	//	- if no IP literal exists, [netip.Addr.IsValid] is false
	Addr() (addr netip.Addr)
	// string representation of this address. May be:
	//	- string value of [Address.Addr]: “1.2.3.4” “::1”
	//	- a domain name “example.com”
	//	- a domain-socket address “/socket” “@socket”
	//	- empty string if domain is empty or Addr is invalid
	fmt.Stringer
}

func NewAddress(domain string) (addr Address) {
	return &address{domain: domain}
}

func NewAddressLiteral(addr netip.Addr) (a Address) {
	return &addressLiteral{addr: addr}
}

type address struct{ domain string }

func (a *address) Addr() (addr netip.Addr) { return }

func (a *address) String() (s string) { return a.domain }

type addressLiteral struct{ addr netip.Addr }

func (a *addressLiteral) Addr() (addr netip.Addr) { return a.addr }

func (a *addressLiteral) String() (s string) { return a.addr.String() }
