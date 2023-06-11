/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// InternetSocket is a unique named type for socket identifiers based on protocol, IP address and possibly port number.
package iana

import (
	"net/netip"

	"github.com/haraldrudell/parl/perrors"
)

// InternetSocket is a unique named type for socket identifiers based on protocol, IP address and possibly port number.
//   - InternetSocket is fmt.Stringer
//   - InternetSocket is ordered
type InternetSocket string

// NewInternetSocket returns an InternetSocket socket identifier based on protocol, IP address and port
func NewInternetSocket(protocol Protocol, addrPort netip.AddrPort) (internetSocket InternetSocket, err error) {
	if !protocol.IsValid() {
		err = perrors.NewPF("protocol not valid")
		return
	}
	if !addrPort.IsValid() {
		err = perrors.NewPF("addrPort not valid")
		return
	}

	// remove possible Zone
	if addrPort.Addr().Zone() != "" {
		addrPort = netip.AddrPortFrom(addrPort.Addr().WithZone(""), addrPort.Port())
	}

	if addrPort.Port() == 0 {
		internetSocket, err = NewInternetSocketNoPort(protocol, addrPort.Addr())
		return
	}

	internetSocket = InternetSocket(protocol.String() + addrPort.String())

	return
}

// NewInternetSocket1 returns an InternetSocket socket identifier based on protocol, IP address and port, panics on error
func NewInternetSocket1(protocol Protocol, addrPort netip.AddrPort) (internetSocket InternetSocket) {
	var err error
	internetSocket, err = NewInternetSocket(protocol, addrPort)
	if err != nil {
		panic(err)
	}
	return
}

// NewInternetSocketNoPort returns an InternetSocket socket identifier based on protocol and IP address
func NewInternetSocketNoPort(protocol Protocol, addr netip.Addr) (internetSocket InternetSocket, err error) {
	if !protocol.IsValid() {
		err = perrors.NewPF("protocol not valid")
		return
	}
	if !addr.IsValid() {
		err = perrors.NewPF("Addr not valid")
		return
	}
	internetSocket = InternetSocket(protocol.String() + addr.String())
	return
}

func (is InternetSocket) IsZeroValue() (isZeroValue bool) {
	return is == ""
}

func (is InternetSocket) String() (s string) {
	return string(is)
}
