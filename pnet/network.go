/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/sets"
)

// the type of Network
//   - NetworkTCP NetworkTCP4 NetworkTCP6
//   - NetworkUDP NetworkUDP4 NetworkUDP6
//   - NetworkIP NetworkIP4 NetworkIP6
//   - NetworkUnix NetworkUnixGram NetworkUnixPacket
type Network string

const (
	NetworkDefault    Network = ""           // default
	NetworkTCP        Network = "tcp"        // tcp IPv4 or IPv6
	NetworkTCP4       Network = "tcp4"       // net network tcp ipv4
	NetworkTCP6       Network = "tcp6"       // tcp IPv6
	NetworkUDP        Network = "udp"        // udp is udp IPv4 or IPv6
	NetworkUDP4       Network = "udp4"       // udp4 is udp IPv4
	NetworkUDP6       Network = "udp6"       // udp6 is udp IPv6
	NetworkIP         Network = "ip"         // ip is IP protocol IPv4 or IPv6 addressing
	NetworkIP4        Network = "ip4"        // ip4 is IP protocol IPv4
	NetworkIP6        Network = "ip6"        // ip6 is IP protocol IPv6
	NetworkUnix       Network = "unix"       // unix is tcp or udp over Unix socket
	NetworkUnixGram   Network = "unixgram"   // unixgram is udp over Unix socket
	NetworkUnixPacket Network = "unixpacket" // unixpacket is tcp over Unix socket
)

// ParseNetwork checks if network is valid
//   - tcp tcp4 tcp6 udp udp4 udp6 ip ip4 ip6 unix unixgram unixpacket
func ParseNetwork(network string) (n Network, err error) {
	n = Network(network)
	if !n.IsValid() {
		err = perrors.Errorf("ParseNetwork: %w", net.UnknownNetworkError(network))
	}
	return
}

var networkSet = sets.NewSet[Network]([]sets.SetElementFull[Network]{
	{ValueV: NetworkTCP, Name: string(NetworkTCP), Full: "tcp IPV4 or IPv6"},
	{ValueV: NetworkTCP4, Name: string(NetworkTCP4), Full: "tcp IPv4"},
	{ValueV: NetworkTCP6, Name: string(NetworkTCP6), Full: "tcp IPv6"},
	{ValueV: NetworkUDP, Name: string(NetworkUDP), Full: "udp IPv4 or IPv6"},
	{ValueV: NetworkUDP4, Name: string(NetworkUDP4), Full: "udp IPv4"},
	{ValueV: NetworkUDP6, Name: string(NetworkUDP6), Full: "udp IPv6"},
	{ValueV: NetworkIP, Name: string(NetworkIP), Full: "IP protocol IPv4 or IPv6 addressing"},
	{ValueV: NetworkIP4, Name: string(NetworkIP4), Full: "IP protocol IPv4"},
	{ValueV: NetworkIP6, Name: string(NetworkIP6), Full: "IP protocol IPv6"},
	{ValueV: NetworkUnix, Name: string(NetworkUnix), Full: "tcp or udp over Unix socket"},
	{ValueV: NetworkUnixGram, Name: string(NetworkUnixGram), Full: "udp over Unix socket"},
	{ValueV: NetworkUnixPacket, Name: string(NetworkUnixPacket), Full: "tcp over Unix socket"},
})

func (t Network) String() (s string) { return networkSet.StringT(t) }

func (t Network) IsValid() (isValid bool) { return networkSet.IsValid(t) }
