/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlnet

import (
	"bytes"
	"context"
	"net"
	"sync"
	"testing"

	"github.com/haraldrudell/parl/error116"
)

func TestNewUdp(t *testing.T) {
	network := "udp"
	address := "127.0.0.1:0"
	maxSize := 0
	var udp *UDP
	packet := []byte{65}
	packetHandler := func(b []byte, oob []byte, flags int, addr *net.UDPAddr) {
		t.Logf("packetHandler: bytes: %d oob: %d flags: %x addr: %s", len(b), len(oob), flags, addr.String())
		if len(oob) > 0 {
			t.Errorf("oob not zero")
			return
		}
		if !bytes.Equal(b, packet) {
			t.Errorf("received packet different from sent packet")
			return
		}
		if flags != 0 {
			t.Errorf("flags not zero")
			return
		}
		t.Logf("shutting down")
		udp.Shutdown() // this will cause errCh to close
	}

	// listen to udp packets
	udp = NewUDP(network, address, packetHandler, maxSize)
	if udp.MaxSize != udpDefaultMaxSize {
		t.Errorf("Bad udp.MaxSize: %d expected %d", udp.MaxSize, udpDefaultMaxSize)
		return
	}

	// error listening thread
	errCh := udp.Listen()
	var errChWg sync.WaitGroup
	errChWg.Add(1)
	go func() {
		defer errChWg.Done()
		t.Log("reading errCh")
		err, ok := <-errCh
		if !ok {
			t.Log("errCh closed")
			return
		}
		t.Error("errCh has error")
		panic(err)
	}()
	isUp, addr := udp.WaitForUp()
	if !isUp {
		t.Logf("bad: WaitForUp: is not up")
		errChWg.Wait()
		t.Errorf("%v", error116.New("WaitForUp: is not up"))
		return
	}
	t.Logf("Listening at %s", addr.String())

	// send udp packet
	var dialer net.Dialer
	ctx := context.Background()
	netConn, err := dialer.DialContext(ctx, addr.Network(), addr.String())
	if err != nil {
		t.Errorf("%v", error116.Errorf("DialContext: '%w'", err))
		return
	}
	t.Logf("sending socket: %s", netConn.LocalAddr().String())
	if n, err := netConn.Write(packet); err != nil {
		t.Errorf("%v", error116.Errorf("netConn.Write: '%w'", err))
		return
	} else {
		t.Logf("wrote %d bytes", n)
	}

	// wait for shutdown from packetHandler
	errChWg.Wait()

	t.Log("Completed successfully")
}

/*
	To continuously listen to

	func net.Listen(network string, address string) (net.Listener, error)
	func net.ListenIP(network string, laddr *net.IPAddr) (*net.IPConn, error)
	func net.ListenMulticastUDP(network string, ifi *net.Interface, gaddr *net.UDPAddr) (*net.UDPConn, error)
	func net.ListenPacket(network string, address string) (net.PacketConn, error)
	func net.ListenTCP(network string, laddr *net.TCPAddr) (*net.TCPListener, error)
	func net.ListenUDP(network string, laddr *net.UDPAddr) (*net.UDPConn, error)
	func net.ListenUnix(network string, laddr *net.UnixAddr) (*net.UnixListener, error)
	func net.ListenUnixgram(network string, laddr *net.UnixAddr) (*net.UnixConn, error)

	Network connection from file:
	func net.FileConn(f *os.File) (c net.Conn, err error)
	func net.FileListener(f *os.File) (ln net.Listener, err error)
	func net.FilePacketConn(f *os.File) (c net.PacketConn, err error)

	Networks:
		"udp"
		"udp4"
		"udp6"
		"unixgram"
		"ip"
		"ip4"
		"ip6"

		var protocols = map[string]int{
			"icmp":      1,
			"igmp":      2,
			"tcp":       6,
			"udp":       17,
			"ipv6-icmp": 58,
		}
		var services = map[string]map[string]int{
			"udp": {
				"domain": 53,
			},
			"tcp": {
				"ftp":    21,
				"ftps":   990,
				"gopher": 70, // ʕ◔ϖ◔ʔ
				"http":   80,
				"https":  443,
				"imap2":  143,
				"imap3":  220,
				"imaps":  993,
				"pop3":   110,
				"pop3s":  995,
				"smtp":   25,
				"ssh":    22,
				"telnet": 23,
*/
