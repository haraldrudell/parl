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
)

const (
	// when Accept exits, errors are sent [SocketListener.close]
	threadExitSendsErrorOnChannel = true
	// when a consumeer invoked Close, errors are not sent [SocketListener.close]
	consumerClose = false
)

// SocketListener is a generic wrapper for [net.Listener]
//   - not intended for direct use, instead use specific implementations like:
//   - — [TCPListener]
//   - C is the type of connection the handler function receives:
//   - — net.Conn, *net.TCPConn, …
//   - panic-handled connection threads
//   - Ch: real-time error channel or collecting errors after close: Err
//   - WaitCh: idempotent waitable thread-terminating Close
//   - SocketListener methods are thread-safe
type SocketListener[C net.Conn] struct {
	// netListener can be type aasserted to a listener for
	// transport socket type
	netListener net.Listener
	// network for listening tcp tcp4 tcp6…
	//	- the type of near and far socket addresses
	//	- network should be compatible wwith transport
	network Network
	// transport indicates what listener implementation netListener
	// can be type asserted to: udp/tcp/ip/unix
	transport SocketTransport
	// stateLock attains integrity by making mutually exclusive:
	//	- [SocketListener.Listen]
	//	- [SocketListener.close]
	//	- [SocketListener.setAcceptState]
	stateLock sync.Mutex
	// state controls the singleton statee cycle: 0 soListening soAccepting soClosing soClosed
	//	- writes behind stateLock for integrity
	state parl.Atomic32[socketState]
	// allows waiting for all pending connections
	connWait sync.WaitGroup
	// allows waiting for accept thread to exit
	acceptWait sync.WaitGroup
	// allows waiting for close complete
	closeWait chan struct{}
	// the channel never closes
	errs parl.ErrSlice
	// cached error from [SocketListener.close]
	closeErr     atomic.Pointer[error]
	threadSource atomic.Pointer[ThreadSource[C]]

	// SocketListener.AcceptConnections

	// the function receiving new connections
	handler func(C)
}

// NewSocketListener returns object listening for socket connections
//   - C is the type of net.Listener the handler function provided to [SocketListener.AcceptConnections]
//   - SocketListener provides asynchronous error handling
//   - handler must invoke net.Conn.Close
//     -
//   - default threading is one virtual thread per connection
//   - [SocketListener.SetThreadSource] allows for any thread model replacing handle
func NewSocketListener[C net.Conn](
	listener net.Listener,
	network Network,
	transport SocketTransport,
	fieldp *SocketListener[C],
	errp *error,
) (socket *SocketListener[C]) {
	var err error
	if listener == nil {
		err = perrors.NewPF("listener cannot be nil")
	} else if !network.IsValid() {
		err = perrors.ErrorfPF("invalid network: %s", network)
	} else if !transport.IsValid() {
		err = perrors.ErrorfPF("invalid transport: %s", transport)
	}
	if err != nil {
		*errp = perrors.AppendError(*errp, err)
		return
	}
	if fieldp == nil {
		// allocation here
		socket = &SocketListener[C]{}
	} else {
		socket = fieldp
	}
	socket.netListener = listener
	socket.network = network
	socket.transport = transport
	socket.closeWait = make(chan struct{})
	return
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

	switch s.state.Load() {
	case soListening, soAccepting:
		err = perrors.NewPF("invoked on listening socket")
		return
	case soClosing, soClosed:
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
	s.state.Store(soListening)

	return
}

// Ch returns a real-time error channel
//   - the channel never closes
//   - unread errors can also be collected using [TCPListener.Err]
func (s *SocketListener[C]) Errs() (errs parl.Errs) { return &s.errs }

// SetThreadSource is a strategy for allocating connect threads
func (s *SocketListener[C]) SetThreadSource(threadSource ThreadSource[C]) {
	s.threadSource.Store(&threadSource)
}

// AcceptConnections is a blocking function handling inbound connections
//   - handler: must be non-nil or a ThreadSource must be active
//   - goodClose true: Accept ended with net.ErrClosed
//   - goodClose false: Accept ended with an unknown error
//   - AcceptConnections:
//   - — accepts connections until the socket is closed by invoking Close
//   - — can only be invoked once and socket state must be Listening
//   - — errors are streamed or collected from [SocketListener.Errs]
//   - handler or ThreadSouce must invoke [net.Conn.Close]
func (s *SocketListener[C]) AcceptConnections(handler func(C)) (goodClose bool) {
	// close the TCP listener
	defer s.close(threadExitSendsErrorOnChannel)
	var err error
	// ensure that the socket is listening ready for accept
	if err = s.setAcceptState(); err != nil {
		// check if it is “use of closed network connection”
		s.errs.AddError(err)
		return
	}
	// cReceiver is optional connection reveiver for connection thread allocation
	var cReceiver ConnectionReceiver[C]
	// shut down any cReceiver and
	//	- await any created connection threads
	//	- and decrement acceptWait from setAcceptState
	defer s.waitForConns(&cReceiver)
	// capture panic
	defer parl.Recover2(func() parl.DA { return parl.A() }, nil, &s.errs)

	s.handler = handler
	for {

		// obtain connection receiver from possible thread source
		//	- receiver is a thread prepared for accept
		if cReceiver, err = s.getReceiver(); err != nil {
			s.errs.AddError(err)
			return // [ThreadSource.Receiver] failed
		} else if cReceiver != nil {
			s.connWait.Add(1)
		} else if handler == nil {
			s.errs.AddError(perrors.NewPF("handler cannot be nil"))
			return // no receiver no handler return
		}

		// block waiting for incoming connection
		var conn net.Conn
		conn, err = s.netListener.Accept()

		// Accept will successfully end with error
		//	- if it is an ending error, AcceeptConnections will return
		//	- any other error is sent on the error channel and continue listening
		if err != nil {
			//	- *net.OpError poll.errNetClosing “accept tcp4 0.0.0.0:57033: use of closed network connection”
			//	- — socket closed upon or during [net.Listener.Accept]
			//	- — package: std/internal/poll symbol: type poll.errNetClosing exported as: var poll.ErrNetClosing
			//	- — re-exported as: net.ErrClosed
			if opError, ok := err.(*net.OpError); ok {
				if goodClose = errors.Is(opError.Err, net.ErrClosed); goodClose {
					// ignore the error
					err = nil
					return // net.ErrClosed “use of closed network connection” return: the socket was closed
				}
			}

			// some accept error: log and keep going
			s.errs.AddError(perrors.ErrorfPF("TCPListener.Accept: %T '%[1]w'", err))
			continue
		}

		// type assert connection: closes conn if assertion fails
		var c C
		if c, err = s.assertConnection(conn); err != nil {
			s.errs.AddError(err)
			return // connection cannot asserted to C return: never happens
		}

		// invoke connection handler
		if cReceiver != nil {
			// isPanic is not used
			var isPanic bool
			s.invokeHandle(c, cReceiver, &isPanic)
		} else {
			s.connWait.Add(1)
			go s.invokeHandler(c)
		}
	}
}

// IsAccept indicates whether the listener is functional and
// accepting incoming connections
func (s *SocketListener[C]) IsAccept() (isAcceptThread bool) { return s.state.Load() == soAccepting }

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
	for _, err := range s.errs.Errors() {
		*errp = perrors.AppendError(*errp, err)
	}
}

// Close ensures the socket is closed
//   - socket guaranteed to be close on return
//   - idempotent panic-free awaitable thread-safe
func (s *SocketListener[C]) Close() (err error) {
	_, err = s.close(consumerClose)
	return
}

// setAcceptState transitions from [soListening] to [soAccepting]
// in critical section
func (s *SocketListener[C]) setAcceptState() (err error) {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	switch s.state.Load() {
	case soListening:
		s.state.Store(soAccepting)
		s.acceptWait.Add(1)
	case 0:
		err = perrors.NewPF("socket not listening")
	case soAccepting:
		err = perrors.NewPF("invoked on accepting socket")
	case soClosing, soClosed:
		err = perrors.NewPF("invoked on closed socket")
	}
	return
}

// close closes [SocketListener.netListener]
//   - only the
func (s *SocketListener[C]) close(sendError bool) (didClose bool, err error) {
	if s.state.Load() == soClosed {
		if ep := s.closeErr.Load(); ep != nil {
			err = *ep
		}
		return // already closed return
	}
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	// select closing invocation
	if didClose = s.state.Load() != soClosed; !didClose {
		if ep := s.closeErr.Load(); ep != nil {
			err = *ep
		}
		return // already closed return
	}

	// execute close
	s.state.Store(soClosing)
	defer close(s.closeWait)
	defer s.state.Store(soClosed)
	defer s.acceptWait.Wait()
	if parl.Close(s.netListener, &err); perrors.Is(&err, "TCPListener.Close %w", err) {
		s.closeErr.Store(&err)
		if sendError {
			s.errs.AddError(err)
		}
	}

	return
}

func (s *SocketListener[C]) invokeHandle(connImpl C, cReceiver ConnectionReceiver[C], isPanic *bool) {
	*isPanic = true

	// TODO 240430: on panic, it is unknown whether the socket was closed
	//	- parl.IdempotentCloser cannot be used, because connImp is an implementation C
	//	- do nothing for now
	cReceiver.Handle(connImpl)

	*isPanic = false
}

// waitForConns shuts down any connection receiver and
// await all connection threads
func (s *SocketListener[C]) waitForConns(cReceiverp *ConnectionReceiver[C]) {
	if cReceiver := *cReceiverp; cReceiver != nil {
		cReceiver.Shutdown()
	}
	// wait for connection goroutines to exit
	s.connWait.Wait()
	// signal accept wait complete
	s.acceptWait.Done()
}

// invokeHandler is a goroutine executing the handler function for a new connection
//   - invokeHandler recovers panic in handler function
func (s *SocketListener[C]) invokeHandler(connImpl C) {
	defer s.connWait.Done()
	defer parl.Recover2(func() parl.DA { return parl.A() }, nil, &s.errs)

	s.handler(connImpl)
}

// obtain handler from possible thread source
func (s *SocketListener[C]) getReceiver() (cReceiver ConnectionReceiver[C], err error) {

	var ts ThreadSource[C]
	if tsp := s.threadSource.Load(); tsp != nil {
		ts = *tsp
	}
	if ts == nil {
		return
	}

	if cReceiver, err = ts.Receiver(&s.connWait, &s.errs); err != nil {
		return // error from [ThreadSource.Receiver]
	} else if cReceiver == nil {
		err = perrors.NewPF("Received nil ConnectionReceiver")
		return // [ThreadSource.Receiver] returned nil
	}

	return // good non-nil return, err nil
}

// type assert connection
func (s *SocketListener[C]) assertConnection(conn net.Conn) (c C, err error) {

	var ok bool
	if c, ok = conn.(C); ok {
		return
	}

	err = perrors.ErrorfPF("connection assertion to %T failed for type %T", c, conn)
	var e error
	parl.Close(conn, &e)
	if e != nil {
		err = perrors.AppendError(err, e)
	}
	return
}
