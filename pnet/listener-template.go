/*
© 2026–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"context"
	"math"
	"net"
	"net/netip"
	"strconv"

	"github.com/haraldrudell/parl/perrors"
)

const (
	// [ListenerTemplate.Network]
	IsUDP = true
)

// ListenerTemplate contains configuration information for a listener,
// either a socket address “example.com:0” or an IPv4 or IPv6 address
// literal “1.2.3.4:5”
//   - port number zero means any ephemeral local port
//   - to be used with [net.ListenConfig.Listen] for tcp
//   - to be used with [net.ListenConfig.ListenPacket] for udp
type ListenerTemplate struct {
	// literal is a value type valid when ListenerTemplate
	// prepresents an address literal “1.2.3.4:5” or “[::1]0”
	literal netip.AddrPort
	// listenAddress is when the value is in domain text form
	// “example.com:123” “/tmp/unixSocket”
	listenAddress string
}

func MakeListenerTemplate(hostPortOrStringLiteral string) (template ListenerTemplate, err error) {

	// address literal string case
	var e error
	if template.literal, e = netip.ParseAddrPort(hostPortOrStringLiteral); e == nil {
		// domainOrStringLiteral was address literal return
		//	- “1.2.3.4:5”
		return
	}

	// host:port string case “example.com:123”
	if _ /*host*/, _ /*portString*/, _ /*port*/, err = template.parseHostPort(hostPortOrStringLiteral); err != nil {
		return
	}
	template.listenAddress = hostPortOrStringLiteral

	return
}

func MakeListenerTemplateFromAddrPort(addrPort netip.AddrPort) (template ListenerTemplate) {
	template.literal = addrPort
	return
}

func MakeListenerTemplateUnix(nonEmptyString string) (template ListenerTemplate) {
	template.listenAddress = nonEmptyString
	return
}

func MakeListenerTemplatePanic(domainOrStringLiteral string) (template ListenerTemplate) {
	var err error
	if template, err = MakeListenerTemplate(domainOrStringLiteral); err != nil {
		panic(err)
	}

	return
}

// IsValid returns error if the listener is uninitialized and therefore
// cannot be used as tcp or udp listener that requires a port value
func (t *ListenerTemplate) IsValid() (err error) {

	if t.literal.IsValid() {
		return
	}
	_ /*host*/, _ /*portString*/, _ /*port*/, err = t.parseHostPort(t.listenAddress)

	return
}

// IsValidUnix returns error if the ListenerTemplate is uninitialized
//   - cannot be used as tcp, udp or Unix socket listener
//   - used to test whether ListenerTemplate can be used to configure a Unix listener
func (t *ListenerTemplate) IsValidUnix() (err error) {
	if t.listenAddress != "" || t.literal.IsValid() {
		return
	}
	err = perrors.NewPF("uninitialized ListenerTemplate")

	return
}

// HasAddrPort returns true if ListenerTemplate contains a binary addr-port literal
//   - can be used for tcp or udp listening by IP address
func (t *ListenerTemplate) HasAddrPort() (hasAddrPort bool) { return t.literal.IsValid() }

// AddrPort returns any addr-port binary literal
//   - if no literal was stored, the value is Invalid
func (t *ListenerTemplate) AddrPort() (addrPort netip.AddrPort) { return t.literal }

func (t *ListenerTemplate) Values() (host, portString string, port uint16) {
	var err error
	if t.literal.IsValid() {
		host = t.literal.Addr().String()
		port = t.literal.Port()
		portString = strconv.Itoa(int(port))
		return
	} else if t.listenAddress == "" {
		return
	} else if host, portString, port, err = t.parseHostPort(t.listenAddress); err == nil {
		return
	}
	host = t.listenAddress

	return
}

func (t *ListenerTemplate) Network(isUdp ...bool) (network Network) {
	if len(isUdp) == 0 || !isUdp[0] {

		// tcp case
		if t.literal.IsValid() {
			if t.literal.Addr().Is4() {
				network = NetworkTCP4
			} else {
				network = NetworkTCP6
			}
		} else {
			network = NetworkTCP
		}
		return
	}
	// it’s udp

	if t.literal.IsValid() {
		if t.literal.Addr().Is4() {
			network = NetworkUDP4
		} else {
			network = NetworkUDP6
		}
	} else {
		network = NetworkUDP
	}

	return
}

func (t *ListenerTemplate) Listen(ctx context.Context, config ...*net.ListenConfig) (tcpListener *net.TCPListener, boundAddress netip.AddrPort, err error) {
	var network, address string
	var cfg *net.ListenConfig
	if network, address, cfg, err = t.prepareListen(tcpYes, config...); err != nil {
		return
	}
	var netListener net.Listener
	var ok bool
	if netListener, err = cfg.Listen(ctx, network, address); perrors.Is(&err, "Listen tcp %w", err) {
		return
		// want [net.TCPListener.AcceptTCP]
	} else if tcpListener, ok = netListener.(*net.TCPListener); !ok {
		err = perrors.ErrorfPF("bad type %T exp *net.TCPListener", netListener)
		return
	} else if boundAddress, err = AddrPortFromAddr(tcpListener.Addr()); err != nil {
		if e := tcpListener.Close(); perrors.IsPF(&e, "Close %w", e) {
			err = perrors.AppendError(err, e)
		}
		tcpListener = nil
	}

	return
}

func (t *ListenerTemplate) ListenPacket(ctx context.Context, config ...*net.ListenConfig) (udpListener *net.UDPConn, boundAddress netip.AddrPort, err error) {
	var network, address string
	var cfg *net.ListenConfig
	if network, address, cfg, err = t.prepareListen(IsUDP, config...); err != nil {
		return
	}
	var netPacketConn net.PacketConn
	var ok bool
	if netPacketConn, err = cfg.ListenPacket(ctx, network, address); perrors.Is(&err, "ListenPacket udp %w", err) {
		return
	} else if udpListener, ok = netPacketConn.(*net.UDPConn); !ok {
		err = perrors.ErrorfPF("bad type %T exp *net.UDPConn", netPacketConn)
	} else if boundAddress, err = AddrPortFromAddr(udpListener.LocalAddr()); err != nil {
		if e := udpListener.Close(); perrors.IsPF(&e, "Close %w", e) {
			err = perrors.AppendError(err, e)
		}
		udpListener = nil
	}

	return
}

func (t *ListenerTemplate) prepareListen(isUdp bool, config ...*net.ListenConfig) (network, address string, cfg *net.ListenConfig, err error) {
	if err = t.IsValid(); err != nil {
		return
	} else if len(config) > 0 {
		cfg = config[0]
	}
	if cfg == nil {
		cfg = &defaultListenConfig
	}
	network = t.Network(isUdp).String()
	address = t.String()

	return
}

func (t *ListenerTemplate) parseHostPort(hostPort string) (host, portString string, port uint16, err error) {

	var h, p string
	var i int
	if h, p, err = net.SplitHostPort(hostPort); perrors.IsPF(&err, "SplitHostPort %w", err) {
		return
	} else if i, err = strconv.Atoi(p); perrors.IsPF(&err, "parsing port number %w", err) {
		return
	} else if i < 0 || i > math.MaxUint16 {
		err = perrors.ErrorfPF("bad port number %d exp %d–%d",
			i, 0, math.MaxUint16,
		)
		return
	}
	host = h
	portString = p
	port = uint16(i)

	return
}

func (t *ListenerTemplate) String() (s string) {
	if t.literal.IsValid() {
		s = t.literal.String()
		return
	}
	s = t.listenAddress

	return
}

const (
	tcpYes = false
)

var (
	defaultListenConfig net.ListenConfig
)
