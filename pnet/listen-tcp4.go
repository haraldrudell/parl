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
type TCPListener struct {
	net.TCPListener // the IPv4 listening socket

	stateLock sync.Mutex
	state     tcpState

	handler    func(net.Conn)
	connWait   sync.WaitGroup     // allows waiting for all pending connections
	acceptWait sync.WaitGroup     // allows waiting for accept thread to exit
	closeWait  chan struct{}      // allows waiting for close complete
	closeErr   error              // cached error from close
	errCh      parl.NBChan[error] // the channel never closes
}

const (
	tcp4 = "tcp4" // net network tcp ipv4
)

// NewListenTCP4 returns object for receiving IPv4 tcp connections
//   - handler must invoke net.Conn.Close
func NewListenTCP4() (socket *TCPListener) { return &TCPListener{closeWait: make(chan struct{})} }

// Listen binds listening to a tcp socket
//   - socketString is host:port "1.2.3.4:80"
//   - — host must be literal IPv4 address 1.2.3.4
//   - — port must be literal port number 0…65534 where 0 means a temporary port
//   - network is always "tcp4"
//   - Listen can be repeatedly invoked until it succeeds
func (s *TCPListener) Listen(socketString string) (err error) {
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

	// resolve near socket address
	var tcpAddr *net.TCPAddr
	if tcpAddr, err = net.ResolveTCPAddr(tcp4, socketString); perrors.Is(&err, "ResolveTCPAddr: '%w'", err) {
		return
	}

	// attempt to listen
	var netTCPListener *net.TCPListener
	if netTCPListener, err = net.ListenTCP(tcp4, tcpAddr); perrors.Is(&err, "ListenTCP: %w", err) {
		return
	}

	// update state to listening
	s.TCPListener = *netTCPListener
	atomic.StoreUint32((*uint32)(&s.state), uint32(tcpListening))

	return
}

// Ch returns a real-time error channel
//   - unread errors can also be collected using [TCPListener.Err]
func (s *TCPListener) Ch() (ch <-chan error) {
	return s.errCh.Ch()
}

// AcceptConnections is a blocking function handling inbound connections
//   - AcceptConnections can only be invoked once
//   - accept of connections continues until Close is invoked
func (s *TCPListener) AcceptConnections(handler func(net.Conn)) {
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
	defer parl.Recover2(parl.Annotation(), nil, s.errCh.AddErrorProc)

	s.handler = handler
	var err error
	var conn net.Conn
	for {

		// block waiting for incoming connection
		if conn, err = s.Accept(); err != nil { // blocking: either a connection or an error
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
func (s *TCPListener) IsAccept() (isAcceptThread bool) {
	return tcpState(atomic.LoadUint32((*uint32)(&s.state))) == tcpAccepting
}

// WaitCh returns a channel that closes when [] completes
//   - ListenTCP4.Close needs to have been invoked for the channel to close
func (s *TCPListener) WaitCh() (closeWait chan struct{}) {
	return s.closeWait
}

func (s *TCPListener) AddrPort() (addrPort netip.AddrPort, err error) {
	addrPort, err = netip.ParseAddrPort(s.Addr().String())
	perrors.IsPF(&err, "netip.ParseAddrPort %w", err)
	return
}

// Err returns all unread errors
//   - errors can also be read using [TCPListener.Ch]
func (s *TCPListener) Err(errp *error) {
	if errp == nil {
		panic(perrors.NewPF("errp cannot be nil"))
	}
	for _, err := range s.errCh.Get() {
		*errp = perrors.AppendError(*errp, err)
	}
}

func (s *TCPListener) Close() (err error) {
	_, err = s.close(false)
	return
}

func (s *TCPListener) setAcceptState() (err error) {
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

func (s *TCPListener) close(sendError bool) (didClose bool, err error) {
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
	if err = s.TCPListener.Close(); perrors.Is(&err, "TCPListener.Close %w", err) {
		s.closeErr = err
		if sendError {
			s.errCh.Send(err)
		}
	}

	return
}

// invokeHandler is a goroutine executing the handler function for a new connection
//   - invokeHandler recovers panics in handler function
func (s *TCPListener) invokeHandler(conn net.Conn) {
	defer s.connWait.Done()
	defer parl.Recover2(parl.Annotation(), nil, s.errCh.AddErrorProc)

	s.handler(conn)
}
