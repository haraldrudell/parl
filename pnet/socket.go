/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"
	"sync"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// Socket embeds net.TCPListener
type Socket struct {
	*net.TCPListener
	wg sync.WaitGroup
}

const (
	tcp4 = "tcp4" // net network tcp ipv4
)

// ListenTCP4 listens on local network interfaces with ipv4 tcp socket: ":8080" or "1.2.3.4:80"
func ListenTCP4(socketString string) (socket *Socket, err error) {
	var s Socket
	var tcpAddr *net.TCPAddr
	if tcpAddr, err = net.ResolveTCPAddr(tcp4, socketString); err != nil {
		err = perrors.Errorf("ResolveTCPAddr: '%w'", err)
		return
	}
	if s.TCPListener, err = net.ListenTCP(tcp4, tcpAddr); err != nil {
		err = perrors.Errorf("ListenTCP: '%w'", err)
		return
	}
	socket = &s
	return
}

// RunHandler handles inbound connections
func (s *Socket) RunHandler(handler func(net.Conn)) (errCh <-chan error) {
	errChan := make(chan error)
	s.wg.Add(1)
	go s.accept(handler, errChan)
	return errChan
}

func (s *Socket) accept(handler func(net.Conn), errCh chan<- error) {
	defer close(errCh)
	defer s.wg.Done()
	defer parl.Recover2(parl.Annotation(), nil, func(e error) { errCh <- e })

	for {
		var err error
		var conn net.Conn
		if conn, err = s.Accept(); err != nil { // blocking: either a connection or an error
			if opError, ok := err.(*net.OpError); ok {
				if opError.Op == "accept" {
					break // ListenTCP4 is closed
				}
			}
			errCh <- perrors.Errorf("TCPListener.Accept: %T '%[1]w'", err) // some error
			continue
		}

		// invoke connection handler
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			handler(conn)
		}()
	}
}

// Wait waits for all connections and the handler thread to exit. ListenTCP4.Close needs to be invoked
func (s *Socket) Wait() {
	s.wg.Wait()
}
