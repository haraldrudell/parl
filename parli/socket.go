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

// Socket is a representation of a network connection
//   - may be inbound or outbound
//   - may be for all interfaces
//   - only exists for IP protocols
//   - may be raw socket
type Socket interface {
	// Local is the near ip address and port
	//	- local IP is 0.0.0.0 or :: for wildcard bind
	//	- port is 0 for 24% sendto udp client sockets, which is also possible for tcp
	//	- port is 0 for non-port protocols such as icmp igmp
	//	- broadcast and multicast udp receive is custom
	Local() (local netip.AddrPort)
	// Remote is the far IP address and port
	//	- 0.0.0.0:0 or :::0 for listening ports
	Remote() (remote netip.AddrPort)
	// Protocol is the IP protocol. Protocol is always present
	Protocol() (protocol iana.Protocol)
	// AddressFamily is the type of address used
	//	- It is undefined 0 for 2% few dns connections
	//	- platforms may have proprietary handling of ip46 ports, ie. listening for both IPv4 and IPv6
	AddressFamily() (addressFamily iana.AddressFamily)
	// InternetSocket is a near-end quasi-unique identifier
	//	- opaque identifier based on IP, port and protocol
	//	- shared ports, sendto udp clients or non-port protocols may not be unique
	InternetSocket() (internetSocket iana.InternetSocket)
	//	Pid is the process identifier for the process using the Socket
	//	- Pid is undefined 0 for 26% tcp ports in LISTEN, TIME_WAIT or CLOSED
	Pid() (pid pids.Pid)
}
