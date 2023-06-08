/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"context"
	"errors"
	"io"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

func TestSocket(t *testing.T) {
	var expAddr = netip.MustParseAddr("127.0.0.1") // localhost: ensure IPv4
	var socketString = expAddr.String() + ":0"     // 0 means ephemeral port

	var colonPos int
	var socket *TCPListener
	var err error
	var addr string
	var port int
	var addrPort netip.AddrPort

	// check socketString
	if colonPos = strings.Index(socketString, ":"); colonPos == -1 {
		t.Fatalf("Bad socketString fixture: %q", socketString)
	}

	// NewTCPListener()
	socket = NewTCPListener()
	if socket == nil {
		t.Fatalf("NewTCPListener nil")
	} else if socket.TCPListener == nil {
		t.Fatalf("socket.TCPListener nil")
	} else if socket.SocketListener.netListener != socket.TCPListener {
		t.Fatalf("socket.TCPListener and socket.SocketListener.netListener different")
	}

	// Listen()
	if err = socket.Listen(socketString); err != nil {
		t.Fatalf("ListenTCP4 error: %+v", err)
	}

	// Addr()
	if addr = socket.Addr().String(); !strings.HasPrefix(addr, socketString[:colonPos+1]) {
		t.Fatalf("Bad socket adress: %q", addr)
	}
	if port, err = strconv.Atoi(strings.TrimPrefix(addr, socketString[:colonPos+1])); err != nil {
		t.Errorf("Bad port number: %q", addr)
	} else if port < 1 || port > 65535 {
		t.Errorf("Bad port numeric: %v", port)
	}

	// AddrPort()
	if addrPort, err = socket.AddrPort(); err != nil {
		t.Errorf("AddrPort err: %s", perrors.Short(err))
	}
	if addrPort.Addr() != expAddr {
		t.Errorf("bad AddrPort addr: %q exp %q", addrPort.Addr(), expAddr)
	}

	// Close()
	if err = socket.Close(); err != nil {
		t.Errorf("socket.Close: '%v'", err)
	}
}

type connectionHandlerFixture struct {
	count int64
}

func (c *connectionHandlerFixture) connFunc(conn *net.TCPConn) {
	if err := conn.Close(); err != nil {
		panic(perrors.Errorf("conn.Close: '%w'", err))
	}
	atomic.AddInt64(&c.count, 1)
}

func (c *connectionHandlerFixture) errorListenerThread(
	socketErrCh <-chan error,
	socketCloseCh <-chan struct{},
	wg *sync.WaitGroup) {
	defer wg.Done()

	var err error
	var ok bool
	for {
		select {
		case <-socketCloseCh:
			return
		case err, ok = <-socketErrCh:
			if !ok {
				panic(perrors.New("socket error channel closed"))
			}
			panic(err)
		}
	}
}

func TestAcceptThread(t *testing.T) {
	var socketString = "127.0.0.1:0" // 0 means ephemeral port
	var fixture connectionHandlerFixture

	var socket *TCPListener
	var err error
	var ctx context.Context = context.Background()
	var addr net.Addr
	var tcpClient net.Dialer
	var netConn net.Conn
	var threadWait sync.WaitGroup

	// set-up socket
	socket = NewTCPListener()
	if err = socket.Listen(socketString); err != nil {
		t.Fatalf("ListenTCP4 error: %+v", err)
	}

	// error listener thread
	threadWait.Add(1)
	go fixture.errorListenerThread(socket.Ch(), socket.WaitCh(), &threadWait)

	// invoke AcceptConnections
	t.Log("socket.AcceptThread…")
	go socket.AcceptConnections(fixture.connFunc)

	// connect to socket
	t.Log("tcpClient.DialContext…")
	addr = socket.Addr()
	if netConn, err = tcpClient.DialContext(ctx, addr.Network(), addr.String()); err != nil {
		t.Fatalf("tcpClient.DialContext: '%v'", err)
	}

	// read from socket
	t.Log("netConn.Read…")
	bytes := make([]byte, 1)
	for {
		var n int
		n, err = netConn.Read(bytes)
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
				break
			} else {
				t.Fatalf("conn.Read: '%v'", err)
			}
		}
		if n != 0 {
			t.Fatalf("conn.Read unexpected bytes: %d", n)
		}
	}

	// close client
	t.Log("netConn.Close…")
	if err := netConn.Close(); err != nil {
		t.Errorf("client Close: '%v'", err)
	}

	// close listener
	t.Log("socket.Close…")
	if err := socket.Close(); err != nil {
		t.Errorf("client Close: '%v'", err)
	}

	t.Logf("socket.Wait… %d", atomic.LoadInt64(&fixture.count))
	<-socket.WaitCh()

	t.Log("error listener Wait…")
	threadWait.Wait()

	t.Log("Completed")
}
