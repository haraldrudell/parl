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
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

func TestSocket(t *testing.T) {
	var socketString = "127.0.0.1:0" // 0 means ephemeral port
	var dummyHandler = func(net.Conn) {}

	var colonPos int
	var socket *TCPListener
	var err error
	var addr string
	var port int

	// check socketString
	if colonPos = strings.Index(socketString, ":"); colonPos == -1 {
		t.Errorf("Bad socketString fixture: %q", socketString)
		t.FailNow()
	}

	if socket, err = NewListenTCP4(socketString, dummyHandler); err != nil {
		t.Errorf("ListenTCP4 error: %+v", err)
		t.FailNow()
	}
	if addr = socket.Addr().String(); !strings.HasPrefix(addr, socketString[:colonPos+1]) {
		t.Errorf("Bad socket adress: %q", addr)
		t.FailNow()
	}
	if port, err = strconv.Atoi(strings.TrimPrefix(addr, socketString[:colonPos+1])); err != nil {
		t.Errorf("Bad port number: %q", addr)
	} else if port < 1 || port > 65535 {
		t.Errorf("Bad port numeric: %v", port)
	}
	if err = socket.Close(); err != nil {
		t.Errorf("socket.Close: '%v'", err)
	}
}

type con struct {
	count int64
}

func (c *con) connFunc(conn net.Conn) {
	if err := conn.Close(); err != nil {
		panic(perrors.Errorf("conn.Close: '%w'", err))
	}
	atomic.AddInt64(&c.count, 1)
}

func (c *con) errListenThread(ch <-chan error, wg *sync.WaitGroup, t *testing.T) {
	defer wg.Done()

	var err error
	var ok bool
	for {
		if err, ok = <-ch; !ok {
			return
		}

		t.Errorf("errCh: %+v", err)
	}
}

func TestAcceptThread(t *testing.T) {
	var socketString = "127.0.0.1:0" // 0 means ephemeral port
	var c con

	var socket *TCPListener
	var err error
	var ctx context.Context = context.Background()
	var addr net.Addr
	var tcpClient net.Dialer
	var netConn net.Conn

	// set-up socket
	if socket, err = NewListenTCP4(socketString, c.connFunc); err != nil {
		t.Errorf("ListenTCP4 error: %+v", err)
		t.FailNow()
	}

	// error listener
	var wg sync.WaitGroup
	wg.Add(1)
	go c.errListenThread(socket.Ch(), &wg, t)

	// launch accept thread
	t.Log("socket.AcceptThread…")
	go socket.AcceptThread()

	// connect to socket
	addr = socket.Addr()
	t.Log("tcpClient.DialContext…")
	if netConn, err = tcpClient.DialContext(ctx, addr.Network(), addr.String()); err != nil {
		t.Errorf("tcpClient.DialContext: '%v'", err)
		t.FailNow()
	}

	// read from socket
	bytes := make([]byte, 1)

	t.Log("netConn.Read…")
	for {
		var n int
		n, err = netConn.Read(bytes)
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
				break
			} else {
				t.Errorf("conn.Read: '%v'", err)
				t.FailNow()
			}
		}
		if n != 0 {
			t.Errorf("conn.Read unexpected bytes: %d", n)
			t.FailNow()
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

	t.Logf("socket.Wait… %d", atomic.LoadInt64(&c.count))
	socket.Wait()

	t.Log("error listener Wait…")
	wg.Wait()

	t.Log("Completed")
}
