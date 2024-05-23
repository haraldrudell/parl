/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phttp

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pnet"
)

const (
	// default port for http: 80, or address for localhost IPv4 or IPv6 port 80
	HttpAddr = ":http"
)

// Http is an http server instance
//   - based on [http.Server]
//   - has listener thread
//   - all errors sent on channel
//   - idempotent deferrable panic-free shutdown
//   - awaitable, observable
type Http struct {
	// used for [http.Server.Listen] invocation
	//	- "tcp", "tcp4", "tcp6", "unix" or "unixpacket"
	Network pnet.Network
	// Close() ListenAndServe() ListenAndServeTLS() RegisterOnShutdown()
	// Serve() ServeTLS() SetKeepAlivesEnabled() Shutdown()
	http.Server
	// real-time server error stream, unbound non-blocking
	//	- [Http.SendErr] invocations
	//   - shutdown error
	ErrCh parl.ErrSlice
	// near socket address, protocol is tcp
	Near netip.AddrPort
	// allows to wait for listen
	//	- when triggered, [Http.Near] is present
	ListenAwaitable parl.Awaitable
	// allows to wait for end of listen
	//	- end of thread launched by [Http.Listen]
	EndListenAwaitable parl.Awaitable
	// the URL router
	serveMux *http.ServeMux
	// whether Listen invocation is allowed
	NoListen atomic.Bool
	// Cancel of listening set-up
	Cancel       atomic.Pointer[context.CancelFunc]
	shutdownOnce sync.Once
}

// NewHttp creates http server for default “localhost:80”
//   - if nearSocket.Addr is invalid, all interfaces for IPv6 if allowed, IPv4 otherwise is used
//   - if nearSocket.Port is zero:
//   - — if network is NetworkDefault: ephemeral port
//   - — otherwise port 80 “:http” is used
//   - for NetworkDefault, NetworkTCP is used
//   - panic for bad Network
//
// Usage:
//
//	var s = NewHttp(netip.AdddrPort{}, pnet.NetworkTCP)
//	s.HandleFunc("/", myHandler)
//	defer s.Shutdown()
//	for err := range s.Listen() {
func NewHttp(nearSocket netip.AddrPort, network pnet.Network) (hp *Http) {
	var hostPort string
	if a := nearSocket.Addr(); a.IsValid() {
		if nearSocket.Port() != 0 || network == pnet.NetworkDefault {
			hostPort = nearSocket.String()
		} else {
			hostPort = a.String() + HttpAddr
		}
	} else {
		hostPort = HttpAddr // default “:http” meaning IPv4 or IPv6 localhost port 80
	}
	switch network {
	case pnet.NetworkDefault:
		network = pnet.NetworkTCP
	case pnet.NetworkTCP, pnet.NetworkTCP4, pnet.NetworkTCP6,
		pnet.NetworkUnix, pnet.NetworkUnixPacket:
	default:
		panic(perrors.ErrorfPF("Bad network: %s allowed: tcp tcp4 tcp6 unix unixpacket", network))
	}
	var serveMux = http.NewServeMux()
	// there is no new-function for [http.Server]
	return &Http{
		Network:  network,
		serveMux: serveMux,
		Server: http.Server{
			// ServeMux matches the URL of each incoming request against a list of registered patterns and calls the handler for the pattern that most closely matches the URL.
			//	- http.Handler is interface { ServeHTTP(ResponseWriter, *Request) }
			Handler: serveMux, // struct
			Addr:    hostPort,
		},
	}
}

const (
	httpShutdownTimeout = 5 * time.Second
)

// HandlerFunc is the signature for URL handlers
type HandlerFunc func(http.ResponseWriter, *http.Request)

// HandleFunc registers a URL-handler for the server
func (s *Http) HandleFunc(pattern string, handler HandlerFunc) {
	s.serveMux.HandleFunc(pattern, handler)
}

// Listen initiates listening and returns the error channel
//   - can only be invoked once or panic
//   - errCh closes on server shutdown
//   - non-blocking, all errors are sent on the error channel
func (s *Http) Listen() (errCh parl.Errs) {
	if !s.NoListen.CompareAndSwap(false, true) {
		panic(perrors.NewPF("multiple invocations"))
	}
	errCh = &s.ErrCh
	// listen is deferred so just launch the thread
	go s.httpListenerThread()
	return
}

// httpListenerThread is gorouitn starting listen and
// waiting for server to terminate
func (s *Http) httpListenerThread() {
	defer s.EndListenAwaitable.Close()
	var err error
	defer parl.Recover(func() parl.DA { return parl.A() }, &err, &s.ErrCh)

	// get near tcp socket listener
	var listener net.Listener
	if listener, err = pnet.Listen(s.Network, s.Server.Addr, &s.Cancel); err != nil {
		return
	}
	defer s.maybeClose(&listener, &err)

	// set Near socket address
	if s.Near, err = pnet.AddrPortFromAddr(listener.Addr()); err != nil {
		return
	}
	s.ListenAwaitable.Close()

	// blocks here until Shutdown or Close
	err = s.Server.Serve(listener)
	listener = nil

	// on regular close, http.ErrServerClosed
	if err == http.ErrServerClosed {
		err = nil // ignore error
		return    // successful return
	}
	err = perrors.Errorf("hp.Server.Serve: ‘%w’", err)
}

// idempotent panic-free shutdown that does not return prior to server shut down
func (s *Http) Shutdown() {
	s.shutdownOnce.Do(s.shutdown)
}

// closes listener if non-nil
func (s *Http) maybeClose(listenerp *net.Listener, errp *error) {
	var listener = *listenerp
	if listener == nil {
		return
	}
	parl.Close(listener, errp)
}

// 5-second shutdown
func (s *Http) shutdown() {
	var wasListen = s.NoListen.Load()
	if !wasListen {
		// prevent further listen
		wasListen = s.NoListen.Swap(true)
	} else {
		if cancelFuncp := s.Cancel.Load(); cancelFuncp != nil {
			(*cancelFuncp)()
		}
	}
	var ctx, cancel = context.WithTimeout(context.Background(), httpShutdownTimeout)
	defer cancel()
	if err := s.Server.Shutdown(ctx); perrors.IsPF(&err, "hp.Server.Shutdown: '%w'", err) {
		s.ErrCh.AddError(err)
	}
	if wasListen {
		// wait for thread to exit
		<-s.EndListenAwaitable.Ch()
	}
	s.ErrCh.EndErrors()
}
