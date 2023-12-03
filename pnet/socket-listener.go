/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"errors"
	"net"
	"net/netip"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/sets"
)

const (
	tcpListening tcpState = iota + 1
	tcpAccepting
	tcpClosing
	tcpClosed
)

type tcpState uint32

const (
	sendErrorOnChannel = true
)

// TCPListener embeds net.TCPListener
//   - panic-handled connection threads
//   - Ch: real-time error channel or collecting errors after close: Err
//   - WaitCh: idempotent waitable thread-terminating Close
//   - C is net.Conn, *net.TCPConn, …
type SocketListener[C net.Conn] struct {
	netListener net.Listener
	network     Network         // network for listening tcp tcp4 tcp6…
	transport   SocketTransport // transport udp/tcp/ip/unix

	stateLock sync.Mutex
	state     tcpState

	handler    func(C)
	connWait   sync.WaitGroup     // allows waiting for all pending connections
	acceptWait sync.WaitGroup     // allows waiting for accept thread to exit
	closeWait  chan struct{}      // allows waiting for close complete
	closeErr   error              // cached error from close
	errCh      parl.NBChan[error] // the channel never closes
}

// NewSocketListener returns object for receiving IPv4 tcp connections
//   - handler must invoke net.Conn.Close
func NewSocketListener[C net.Conn](
	listener net.Listener,
	network Network,
	transport SocketTransport,
) (socket *SocketListener[C]) {
	if listener == nil {
		panic(perrors.NewPF("listener cannot be nil"))
	} else if !network.IsValid() {
		panic(perrors.ErrorfPF("invalid network: %s", network))
	} else if !transport.IsValid() {
		panic(perrors.ErrorfPF("invalid transport: %s", transport))
	}
	return &SocketListener[C]{
		netListener: listener,
		network:     network,
		transport:   transport,
		closeWait:   make(chan struct{}),
	}
}

// Listen binds listening to a near socket
//   - socketString is host:port "1.2.3.4:80" "wikipedia.com:443" "/some/unix/socket"
//   - — for TCP UDP IP host must resolve to an assigned near IP address
//   - — — if host is blank, it is for localhost
//   - — — to avoid DNS resolution host should be blank or literal IP address "1.2.3.4:0"
//   - — for TCP UDP port must be literal port number 0…65534 where 0 means a temporary port
//   - Listen can be repeatedly invoked until it succeeds
func (s *SocketListener[C]) Listen(socketString string) (err error) {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()
	switch tcpState(atomic.LoadUint32((*uint32)(&s.state))) {
	case tcpListening, tcpAccepting:
		err = perrors.NewPF("invoked on listening socket")
		return
	case tcpClosing, tcpClosed:
		err = perrors.NewPF("invoked on closed socket")
		return
	}

	switch s.transport {
	case TransportTCP:
		// resolve near socket address
		var tcpAddr *net.TCPAddr
		if tcpAddr, err = net.ResolveTCPAddr(s.network.String(), socketString); perrors.Is(&err, "ResolveTCPAddr: '%w'", err) {
			return
		}

		// attempt to listen
		var netTCPListener *net.TCPListener
		if netTCPListener, err = net.ListenTCP(s.network.String(), tcpAddr); perrors.Is(&err, "ListenTCP: %w", err) {
			return
		}

		// copy socket to TCPListener storage
		var listenerp = s.netListener.(*net.TCPListener)
		*listenerp = *netTCPListener
	default:
		err = perrors.ErrorfPF("unimplemented transport: %s", s.transport)
		return
	}

	// update state to listening
	atomic.StoreUint32((*uint32)(&s.state), uint32(tcpListening))

	return
}

// Ch returns a real-time error channel
//   - unread errors can also be collected using [TCPListener.Err]
func (s *SocketListener[C]) Ch() (ch <-chan error) {
	return s.errCh.Ch()
}

// AcceptConnections is a blocking function handling inbound connections
//   - AcceptConnections can only be invoked once
//   - accept of connections continues until Close is invoked
//   - handler must invoke net.Conn.Close
func (s *SocketListener[C]) AcceptConnections(handler func(C)) {
	defer s.close(sendErrorOnChannel)
	if handler == nil {
		s.errCh.Send(perrors.NewPF("handler cannot be nil"))
		return
	}
	if err := s.setAcceptState(); err != nil {
		s.errCh.Send(err)
		return
	}
	defer s.acceptWait.Done() // indicate accept thread exited
	defer s.connWait.Wait()   // wait for connection goroutines
	defer parl.Recover2(func() parl.DA { return parl.A() }, nil, s.errCh.AddErrorProc)

	s.handler = handler
	var err error
	var conn net.Conn
	for {

		// block waiting for incoming connection
		if conn, err = s.netListener.Accept(); err != nil { // blocking: either a connection or an error
			if opError, ok := err.(*net.OpError); ok {
				if errors.Is(opError.Err, net.ErrClosed) {
					return // use of closed: assume shutdown: ListenTCP4 is closed
				}
			}
			s.errCh.Send(perrors.ErrorfPF("TCPListener.Accept: %T '%[1]w'", err)) // some error
			continue
		}

		// invoke connection handler
		s.connWait.Add(1)
		go s.invokeHandler(conn)
	}
}

// IsAccept indicates whether the listener is functional and
// accepting incoming connections
func (s *SocketListener[C]) IsAccept() (isAcceptThread bool) {
	return tcpState(atomic.LoadUint32((*uint32)(&s.state))) == tcpAccepting
}

// WaitCh returns a channel that closes when [] completes
//   - ListenTCP4.Close needs to have been invoked for the channel to close
func (s *SocketListener[C]) WaitCh() (closeWait chan struct{}) {
	return s.closeWait
}

func (s *SocketListener[C]) AddrPort() (addrPort netip.AddrPort, err error) {
	var netAddr = s.netListener.Addr()
	switch a := netAddr.(type) {
	case *net.TCPAddr:
		addrPort = a.AddrPort()
	case *net.UDPAddr:
		addrPort = a.AddrPort()
	case *net.UnixAddr:
		return // unix sockets do not have address or port
	case *net.IPAddr:
		var addr netip.Addr
		if addr, err = netip.ParseAddr(a.String()); perrors.IsPF(&err, "netip.ParseAddr %w", err) {
			return
		}
		addrPort = netip.AddrPortFrom(addr, 0)
	default:
		addrPort, err = netip.ParseAddrPort(netAddr.String())
		perrors.IsPF(&err, "netip.ParseAddrPort %w", err)
	}
	return
}

// Err returns all unread errors
//   - errors can also be read using [TCPListener.Ch]
func (s *SocketListener[C]) Err(errp *error) {
	if errp == nil {
		panic(perrors.NewPF("errp cannot be nil"))
	}
	for _, err := range s.errCh.Get() {
		*errp = perrors.AppendError(*errp, err)
	}
}

func (s *SocketListener[C]) Close() (err error) {
	_, err = s.close(false)
	return
}

func (s *SocketListener[C]) setAcceptState() (err error) {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	switch tcpState(atomic.LoadUint32((*uint32)(&s.state))) {
	case tcpListening:
		atomic.StoreUint32((*uint32)(&s.state), uint32(tcpAccepting))
		s.acceptWait.Add(1)
	case 0:
		err = perrors.NewPF("socket not listening")
	case tcpAccepting:
		err = perrors.NewPF("invoked on accepting socket")
	case tcpClosing, tcpClosed:
		err = perrors.NewPF("invoked on closed socket")
	}
	return
}

func (s *SocketListener[C]) close(sendError bool) (didClose bool, err error) {
	if tcpState(atomic.LoadUint32((*uint32)(&s.state))) == tcpClosed {
		err = s.closeErr
		return // already closed return
	}
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	// select closing invocation
	if didClose = tcpState(atomic.LoadUint32((*uint32)(&s.state))) != tcpClosed; !didClose {
		err = s.closeErr
		return // already closed return
	}

	// execute close
	atomic.StoreUint32((*uint32)(&s.state), uint32(tcpClosing))
	defer close(s.closeWait)
	defer atomic.StoreUint32((*uint32)(&s.state), uint32(tcpClosed))
	defer s.acceptWait.Wait()
	if parl.Close(s.netListener, &err); perrors.Is(&err, "TCPListener.Close %w", err) {
		s.closeErr = err
		if sendError {
			s.errCh.Send(err)
		}
	}

	return
}

// invokeHandler is a goroutine executing the handler function for a new connection
//   - invokeHandler recovers panics in handler function
func (s *SocketListener[C]) invokeHandler(conn net.Conn) {
	defer s.connWait.Done()
	defer parl.Recover2(func() parl.DA { return parl.A() }, nil, s.errCh.AddErrorProc)

	var c C
	var ok bool
	if c, ok = conn.(C); !ok {
		s.errCh.Send(perrors.ErrorfPF("connection assertion to %T failed for type %T", c, conn))
		return
	}

	s.handler(c)
}

const (
	NetworkTCP  = "tcp"
	NetworkTCP4 = "tcp4" // net network tcp ipv4
	NetworkTCP6 = "tcp6"
)

type Network string

func (t Network) String() (s string) {
	return networkSet.StringT(t)
}

func (t Network) IsValid() (isValid bool) {
	return networkSet.IsValid(t)
}

var networkSet = sets.NewSet[Network]([]sets.SetElement[Network]{
	{ValueV: NetworkTCP, Name: "tcp"},
	{ValueV: NetworkTCP4, Name: "tcp4"},
	{ValueV: NetworkTCP6, Name: "tcp6"},
})

const (
	TransportTCP = iota + 1
	TransportUDP
	TransportIP
	TransportUnix
)

type SocketTransport uint8

func (t SocketTransport) String() (s string) {
	return transportSet.StringT(t)
}

func (t SocketTransport) IsValid() (isValid bool) {
	return transportSet.IsValid(t)
}

var transportSet = sets.NewSet[SocketTransport]([]sets.SetElement[SocketTransport]{
	{ValueV: TransportTCP, Name: "tcp"},
})
