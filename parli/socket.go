/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parli

import (
	"net/netip"

	"github.com/haraldrudell/parl/iana"
	"github.com/haraldrudell/parl/pids"
)

type Socket interface {
	Local() (local netip.AddrPort)
	Remote() (remote netip.AddrPort)
	Protocol() (protocol iana.Protocol)
	AddressFamily() (addressFamily iana.AddressFamily)
	InternetSocket() (internetSocket iana.InternetSocket)
	Pid() (pid pids.Pid)
}
