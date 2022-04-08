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
	"sync/atomic"
	"testing"

	"github.com/haraldrudell/parl"
)

func TestSocket(t *testing.T) {
	socketString := "127.0.0.1:0" // 0 means ephemeral port
	colonPos := strings.Index(socketString, ":")
	if colonPos == -1 {
		t.Errorf("Bad socketString fixture: %q", socketString)
		return
	}
	socket, err := ListenTCP4(socketString)
	if err != nil {
		t.Errorf("ListenTCP4 error: %+v", err)
		return
	}
	addr := socket.Addr().String()
	if !strings.HasPrefix(addr, socketString[:colonPos+1]) {
		t.Errorf("Bad socket adress: %q", addr)
		return
	}
	if port, err := strconv.Atoi(strings.TrimPrefix(addr, socketString[:colonPos+1])); err != nil {
		t.Errorf("Bad port number: %q", addr)
		return
	} else if port < 1 || port > 65535 {
		t.Errorf("Bad port numeric: %v", port)
		return
	}
	if err := socket.Close(); err != nil {
		t.Errorf("socket.Close: '%v'", err)
		return
	}
}

func TestRunHandler(t *testing.T) {

	// set up socket
	socketString := "127.0.0.1:0" // 0 means ephemeral port
	socket, err := ListenTCP4(socketString)
	if err != nil {
		t.Errorf("ListenTCP4 error: %+v", err)
		return
	}

	// connection listener
	var count int64
	connFunc := func(conn net.Conn) {
		if err := conn.Close(); err != nil {
			panic(parl.Errorf("conn.Close: '%w'", err))
		}
		atomic.AddInt64(&count, 1)
	}
	errCh := socket.RunHandler(connFunc)
	go func() {
		for {
			if err, ok := <-errCh; !ok {
				break
			} else {
				t.Logf("errCh: %+v", err)
			}
		}
	}()

	// connect to socket
	ctx := context.Background()
	addr := socket.Addr()
	var tcpClient net.Dialer
	conn, err := tcpClient.DialContext(ctx, addr.Network(), addr.String())
	if err != nil {
		t.Errorf("tcpClient.DialContext: '%v'", err)
		return
	}

	// read from socket
	bytes := make([]byte, 1)
	for {
		n, err := conn.Read(bytes)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				t.Errorf("conn.Read: '%v'", err)
				return
			}
		}
		if n != 0 {
			t.Errorf("conn.Read unexpected bytes: %d", n)
			return
		}
	}

	// close client
	if err := conn.Close(); err != nil {
		t.Errorf("client Close: '%v'", err)
		return
	}

	// close listener
	if err := socket.Close(); err != nil {
		t.Errorf("client Close: '%v'", err)
		return
	}

	t.Logf("socket.Wait… %d", atomic.LoadInt64(&count))
	socket.Wait()
}
