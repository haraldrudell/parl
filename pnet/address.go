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
//   - created by [NewAddress] or [NewAddressLiteral]
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

// NewAddress creates an [Address] value based on domain-name
//   - domain: “example.com”
//   - fieldp: optional stack-allocation
//   - —
//   - Address is used with legacy net functions.
//   - [Address.String] provides a string representation
//   - [Address.Addr] returns [netip.Addr] as applicable
func NewAddress(domain string, fieldp ...*AddressDomain) (addr Address) {

	var a *AddressDomain
	if len(fieldp) > 0 {
		a = fieldp[0]
	}
	if a == nil {
		a = &AddressDomain{}
	}
	a.domain = domain
	addr = a

	return
}

func NewAddressLiteral(addr netip.Addr) (a Address) {
	return &AddressLiteral{addr: addr}
}

type AddressDomain struct{ domain string }

func (a *AddressDomain) Addr() (addr netip.Addr) { return }

func (a *AddressDomain) String() (s string) { return a.domain }

type AddressLiteral struct{ addr netip.Addr }

func (a *AddressLiteral) Addr() (addr netip.Addr) { return a.addr }

func (a *AddressLiteral) String() (s string) { return a.addr.String() }
