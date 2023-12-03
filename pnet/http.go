/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"context"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

type Http struct {
	Network string // "tcp", "tcp4", "tcp6", "unix" or "unixpacket"
	http.Server
	ListenInvoked atomic.Bool
	ReadyWg       sync.WaitGroup
	ErrCh         chan<- error
	ErrChMutex    sync.Mutex
	ErrChClosed   atomic.Bool
	net.Addr      // interface
	IsListening   atomic.Bool
	IsShutdown    atomic.Bool
}

// NewHttp creates http server host is host:port, default ":http"
func NewHttp(host, network string) (hp *Http) {
	if host == "" {
		host = httpAddr
	}
	if network == "" {
		network = TCPNetwork
	}
	h := Http{
		Network: network,
		Server: http.Server{
			Handler: http.NewServeMux(), // struct
			Addr:    host,
		},
	}
	return &h
}

const (
	httpAddr            = ":http"
	TCPNetwork          = "tcp"
	httpShutdownTimeout = 5 * time.Second
)

type HandlerFunc func(http.ResponseWriter, *http.Request)

func (hp *Http) HandleFunc(pattern string, handler HandlerFunc) {
	var httpServeMux *http.ServeMux
	var ok bool
	if httpServeMux, ok = hp.Handler.(*http.ServeMux); !ok {
		panic(perrors.Errorf("server.Handler not *http.ServeMux: %T", hp.Handler))
	}
	httpServeMux.HandleFunc(pattern, handler)
}

func (hp *Http) Listen() (errCh <-chan error) {
	errCh = hp.SubListen()
	go hp.listenerThread()
	return
}

func (hp *Http) SubListen() (errCh <-chan error) {
	hp.ReadyWg.Add(1)
	defer hp.ReadyWg.Done()
	if !hp.ListenInvoked.CompareAndSwap(false, true) {
		panic(perrors.New("multiple http.Run invocations"))
	}
	errChan := make(chan error)
	errCh = errChan
	hp.ErrCh = errChan
	hp.ReadyWg.Add(1)
	return
}

func (hp *Http) listenerThread() {
	defer hp.CloseErr()
	defer parl.Recover(func() parl.DA { return parl.A() }, nil, hp.SendErr)
	var didReadyWg bool
	defer func() {
		if !didReadyWg {
			hp.ReadyWg.Done()
		}
	}()

	listener, err := hp.Listener()
	if err != nil {
		return
	}
	hp.ReadyWg.Done()
	didReadyWg = true
	hp.IsListening.Store(true)

	if err := hp.Server.Serve(listener); err != nil { // blocking until Shutdown or Close
		if err != http.ErrServerClosed {
			hp.SendErr(perrors.Errorf("hp.Server.Serve: '%w'", err))
			return
		}
	}
}

func (hp *Http) Listener() (listener net.Listener, err error) {
	srv := &hp.Server
	listener, err = net.Listen(hp.Network, srv.Addr)
	if err != nil {
		hp.SendErr(perrors.Errorf("net.Listen %s %s: '%w'", hp.Network, srv.Addr, err))
		return
	}
	hp.Addr = listener.Addr()
	return
}

func (hp *Http) WaitForUp() (isUp bool, addr net.Addr) {
	if !hp.ListenInvoked.Load() {
		return // Listen has not been invoked
	}
	hp.ReadyWg.Wait()
	if isUp = hp.IsListening.Load(); isUp {
		addr = hp.Addr
	}
	return
}

func (hp *Http) SendErr(err error) {
	hp.ErrChMutex.Lock()
	defer hp.ErrChMutex.Unlock()
	if !hp.ErrChClosed.Load() {
		hp.ErrCh <- err
	}
}

func (hp *Http) CloseErr() {
	if hp.ErrChClosed.CompareAndSwap(false, true) {
		hp.ErrChMutex.Lock()
		close(hp.ErrCh)
		hp.ErrChMutex.Unlock()
	}
}

func (hp *Http) Shutdown() {
	if !hp.IsShutdown.CompareAndSwap(false, true) {
		return // already shutdown
	}
	ctx, cancel := context.WithTimeout(context.Background(), httpShutdownTimeout)
	defer cancel()
	if err := hp.Server.Shutdown(ctx); err != nil {
		hp.SendErr(perrors.Errorf("hp.Server.Shutdown: '%w'", err))
	}
}
