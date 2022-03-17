/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlnet

import (
	"net"
	"sync"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/error116"
)

type UDP struct {
	Network        string
	F              UDPFunc
	MaxSize        int
	net.UDPAddr    // struct IP Port Zone
	ListenInvoked  parl.AtomicBool
	StartingListen sync.WaitGroup
	ErrCh          chan<- error
	IsListening    parl.AtomicBool
	NetUDPConn     *net.UDPConn
	connMutex      sync.RWMutex
	Addr           net.Addr
	IsShutdown     parl.AtomicBool
}

type UDPFunc func(b []byte, oob []byte, flags int, addr *net.UDPAddr)

// NewUDP network: "udp" "udp4" "udp6" address: "host:port"
func NewUDP(network, address string, udpFunc UDPFunc, maxSize int) (udp *UDP) {
	if maxSize < 1 {
		maxSize = udpDefaultMaxSize
	}
	if netUDPAddr, err := net.ResolveUDPAddr(network, address); err != nil {
		panic(error116.Errorf("net.ResolveUDPAddr: '%w'", err))
	} else {
		udp = &UDP{UDPAddr: *netUDPAddr, Network: network, F: udpFunc, MaxSize: maxSize}
	}
	return
}

const (
	udpDefaultMaxSize = 65507 // max for ipv4
	oobSize           = 40
	netReadOperation  = "read"
	useOfClosed       = "use of closed network connection"
)

func (udp *UDP) Listen() (errCh <-chan error) {
	udp.StartingListen.Add(1)
	if !udp.ListenInvoked.Set() {
		udp.StartingListen.Done()
		panic(error116.New("multiple udp.Listen invocations"))
	}
	if udp.IsShutdown.IsTrue() {
		udp.StartingListen.Done()
		panic(error116.New("udp.Listen after Shutdown"))
	}
	errChan := make(chan error)
	errCh = errChan
	udp.ErrCh = errChan
	go udp.listenThread()
	return
}

func (udp *UDP) listenThread() {
	errCh := udp.ErrCh
	defer close(errCh)
	var FInvocations sync.WaitGroup
	defer FInvocations.Wait()
	var startingDone bool
	defer func() {
		if !startingDone {
			udp.StartingListen.Done()
		}
	}()
	defer parl.Recover2(parl.Annotation(), nil, func(e error) { errCh <- e }) // capture panics

	// listen
	var netUDPConn *net.UDPConn // represents a network file descriptor
	var err error
	if netUDPConn, err = net.ListenUDP(udp.Network, &udp.UDPAddr); err != nil {
		errCh <- error116.Errorf("net.ListenUDP: '%w'", err)
		return
	}
	if udp.setConn(netUDPConn) { // isShutdown
		if err = netUDPConn.Close(); err != nil {
			errCh <- error116.Errorf("netUDPConn.Close: '%w'", err)
		}
		return
	}
	udp.Addr = netUDPConn.LocalAddr()
	udp.IsListening.Set()
	udp.StartingListen.Done()
	startingDone = true
	defer func() {
		if !udp.IsShutdown.IsTrue() {
			if err := netUDPConn.Close(); err != nil {
				errCh <- err
			}
		}
		udp.IsListening.Clear()
	}()

	// read datagrams
	for {
		b := make([]byte, udp.MaxSize)
		oob := make([]byte, oobSize)
		var n int
		var oobn int
		var flags int
		var addr *net.UDPAddr
		var err error
		n, oobn, flags, addr, err = netUDPConn.ReadMsgUDP(b, oob)
		if err != nil {
			if udp.IsShutdown.IsTrue() && udp.isClosedErr(err) {
				return // we are shutdown
			}
			errCh <- error116.Errorf("ReadMsgUDP: '%w'", err)
			return
		}
		FInvocations.Add(1)
		go func() {
			defer FInvocations.Done()
			udp.F(b[:n], oob[:oobn], flags, addr)
		}()
	}
}

func (udp *UDP) WaitForUp() (isUp bool, addr net.Addr) {
	if !udp.ListenInvoked.IsTrue() {
		return // Listen has not been invoked
	}
	udp.StartingListen.Wait()
	if isUp = udp.IsListening.IsTrue(); isUp {
		addr = udp.Addr
	}
	return
}

func (udp *UDP) isClosedErr(err error) (isClose bool) {
	// read udp 127.0.0.1:50050: use of closed network connection
	// &net.OpError{Op:"read", Net:"udp", Source:(*net.UDPAddr)(0xc00007a030), Addr:net.Addr(nil), Err:(*errors.errorString)(0xc000098160)}
	opErr, ok := err.(*net.OpError)
	if !ok {
		return
	}
	if opErr.Op != netReadOperation {
		return
	}
	e := opErr.Err
	// &errors.errorString{s:"use of closed network connection"}
	if e.Error() != useOfClosed {
		return // some other error
	}
	isClose = true
	return
}

func (udp *UDP) setConn(conn *net.UDPConn) (isShutdown bool) {
	udp.connMutex.Lock()
	defer udp.connMutex.Unlock()
	isShutdown = udp.IsShutdown.IsTrue()
	if !isShutdown {
		udp.NetUDPConn = conn
	}
	return
}

func (udp *UDP) Shutdown() {
	udp.connMutex.RLock()
	defer udp.connMutex.RUnlock()
	if !udp.IsShutdown.Set() {
		return // it was already shutdown
	}
	conn := udp.NetUDPConn
	if conn == nil {
		return // no need to close connection
	}
	err := conn.Close()
	if err == nil {
		return // conn successfully closed
	}
	udp.ErrCh <- err
}
