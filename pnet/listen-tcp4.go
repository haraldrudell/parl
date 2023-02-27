/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// TCPListener embeds net.TCPListener
//   - NewListenTCP4 initializes
//   - go AcceptThread handles inbound connections
//   - Ch provides the error channel
//   - Close closes TCPListener
//   - Wait awaits last handler exit
type TCPListener struct {

	// read-only fields

	net.TCPListener
	handler func(net.Conn)

	// fields with thread-safe access

	hasAcceptThread atomic.Bool
	handlerWait     sync.WaitGroup
	errs            parl.NBChan[error]
}

const (
	tcp4 = "tcp4" // net network tcp ipv4
)

// NewListenTCP4 returns a tcp listener on a local IPv4 network interface
//   - socketString is host:port "1.2.3.4:80"
//   - — host must be literal IPv4 address 1.2.3.4
//   - — port must be literal port number 0…65534 where 0 means a temporary port
//   - network is always "tcp4"
//   - handler must invoke net.Conn.Close
func NewListenTCP4(socketString string, handler func(net.Conn)) (socket *TCPListener, err error) {
	if handler == nil {
		err = perrors.NewPF("handler cannot be nil")
		return
	}
	var tcpAddr *net.TCPAddr
	var netTCPListener *net.TCPListener
	if tcpAddr, err = net.ResolveTCPAddr(tcp4, socketString); perrors.Is(&err, "ResolveTCPAddr: '%w'", err) {
		return
	}
	if netTCPListener, err = net.ListenTCP(tcp4, tcpAddr); perrors.Is(&err, "ListenTCP: %w", err) {
		return
	}
	socket = &TCPListener{TCPListener: *netTCPListener, handler: handler}
	return
}

func (s *TCPListener) Ch() (ch <-chan error) {
	return s.errs.Ch()
}

// AcceptThread is a goroutine handling inbound connections
func (s *TCPListener) AcceptThread() {
	if s.handler == nil {
		panic(perrors.NewPF("handler cannot be nil"))
	} else if s.hasAcceptThread.Swap(true) {
		panic(perrors.NewPF("AcceptThread invoked more than once"))
	}
	defer s.errs.Close()
	defer s.handlerWait.Wait()
	defer parl.Recover2(parl.Annotation(), nil, s.errs.AddErrorProc)

	var err error
	var conn net.Conn
	for {

		// block waiting for incoming connection
		if conn, err = s.Accept(); err != nil { // blocking: either a connection or an error
			if opError, ok := err.(*net.OpError); ok {
				if opError.Op == "accept" {
					return // ListenTCP4 is closed
				}
			}
			s.errs.Send(perrors.ErrorfPF("TCPListener.Accept: %T '%[1]w'", err)) // some error
			continue
		}

		// invoke connection handler
		s.handlerWait.Add(1)
		go s.invokeHandler(conn)
	}
}

// IsAcceptThread indicates whether the listener is functional and
// accepting incoming connections
func (s *TCPListener) IsAcceptThread() (isAcceptThread bool) {
	return s.hasAcceptThread.Load() && !s.errs.IsClosed()
}

// Wait waits for all connections and the handler thread to exit. ListenTCP4.Close needs to be invoked
func (s *TCPListener) Wait() {
	s.errs.WaitForClose()
}

func (s *TCPListener) Err(errp *error) {
	if errp == nil {
		panic(perrors.NewPF("errp cannot be nil"))
	}
	for _, err := range s.errs.Get() {
		*errp = perrors.AppendError(*errp, err)
	}
}

// invokeHandler is a goroutine executing the handler funciton for a connection
func (s *TCPListener) invokeHandler(conn net.Conn) {
	defer s.handlerWait.Done()
	defer parl.Recover2(parl.Annotation(), nil, s.errs.AddErrorProc)

	s.handler(conn)
}
