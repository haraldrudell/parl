/*
© 2025–present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package phttp

import (
	"context"
	"crypto"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"sync/atomic"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/phttp/phlib"
	"github.com/haraldrudell/parl/pnet"
)

const (
	DefaultCloseTimeout = 3 * time.Second
	NoParLimit          = 0
	DefaultParallelism  = 512
)

// needs http server serving protocol https
//   - creating http.Server instead of using eg. [http.ListenAndServe] has
//     benefits:
//   - — graceful shutdown is possible
//   - — prevents logging to system logger
//   - — can use TLS and self-signed certificates
//   - — can set timeouts
//   - — can use a base context for cancellation and
//     providing server-scoped values to requests
//   - it is easier to separate Listen and Serve methods rather than:
//   - — making Listen events awaitable and
//   - — tracking near addresses for multiple listeners
//   - the [http.Server.Serve] is difficult due to unbound parallelism but cannot be overridden
//   - — therefore, server-scoped bound parallelism in listeners
//   - — at 1,024 goroutines, the Go runtime has significant temporary memory leaks
//   - — default is therefore 512 current or long-duration connections
//   - [Server] compared to [http.Server]:
//   - — [http.Server.ListenAndServe] http is not provided, only https
//   - — [http.Server.ListenAndServeTLS] [http.Server.ServeTLS] provided by
//     [Server.Listen] [Server.Serve]
//   - — [http.Server.RegisterOnShutdown] is [Server.RegisterOnShutdown]
//   - — [http.Server.Serve] http is not provided, only https
//   - — [http.Server.SetKeepAlivesEnabled] not provided, they are enabled by default
//   - — [http.Server.Shutdown] provided by [Server.Close]
//   - — [http.Server.Close] provided by [Server.CloseNow]
//
// http.conn cannot be recreated due to use of vendor packages
// and it’s 600-lines with several references to unexported identifiers
//   - [http.Server.Serve] is the only place where [net.Listener.Accept] is invoked
//   - therefore, a new goroutine will be created for every new connection
//   - [netutil.LimitListener] can limit parallelism in [net.Listener.Accept]
//     does only limit per listener in isolation, not server-wide
//   - parallelism can be limited at [http.Server.Handler] but then unlimited threads
//     may already be created in the [http.Server.Serve] Accept loop
//   - each created goroutine created running http.conn.serve may serve many requests
//     using http/1.1 keep-alive or http/2 multiplexing
//   - therefore limit in [net.Listener.Accept] to 1,024 for any computer with 100,000
//     for custom-configured servers
//
// possibilities for http listen in Go:
//   - two options: [http.Server.ListenAndServe] and the simpler [http.ListenAndServe]
//   - [http.HandleFunc] can be used to provide a handler function that responds
//     to specific URLs from a process-scoped default http handler that is
//     activated by [http.ListenAndServe]
//   - a separate URL-space multiplexer can be created by [http.NewServerMux],
//     [http.ServeMux.HandleFunc] then provided to [http.ListenAndServe]
//   - the mux can take a struct implementing [http.Handler]
//   - for https and other configuration, a separate [http.Server] instance is used.
//     The field [http.Server.Handler] conatins the URL-space multiplexer.
//     [http.Server.ListenAndServe] starts the server
type Server struct {
	// httpServer contains server cofiguration
	//	- has non-pointer locks
	//	- must be heap
	httpServer http.Server
	// closeTimeout is how long graceful shutdown is in [CDavServer.Close]
	closeTimeout time.Duration
	// true when Close or Shutdown have been invoked
	inShutdown atomic.Bool
	// threadLimiter is an optional ticketer limiting
	// threads across all listeners
	threadLimiter parl.Moderate
	// log unused 250928
	log parl.PrintfFunc
	// ctx is base context for requests
	ctx context.Context
}

// NewServerHandler creates a http server with optional log
//   - handler: ServeHTTP method, typically returned by [http.NewServeMux]
//   - log: optional log channeling [http.Server] logging, otherwise to standard error
//   - —
//   - limited parallelism 512 goroutines, 3 s graceful close, default [http.Server] struct
//   - [Server.Listen] listens using in-memory TLS credentials and socket address
//   - [Server.Serve] listens to any number of sockets returned by Listen
//   - [Server.Close] 3 s graceful shutdown
func NewServerHandler(handler http.Handler, log ...parl.PrintfFunc) (c *Server) {
	var logX parl.PrintfFunc
	if len(log) > 0 {
		logX = log[0]
	}
	if logX == nil {
		logX = parl.Log
	}
	c = NewServer(
		DefaultParallelism, context.Background(), DefaultCloseTimeout, logX, &http.Server{
			Handler: handler,
		})
	return
}

// NewServer
//   - parallelism: maximum number of concurrent connections
//   - — [NoParLimit] 0: unlimited
//   - — [DefaultParallelism]: 512 connections/threads which is reasonable
//     on a 2025 laptop
//   - ctx: base context for request-threads
//   - closeTimeout: [DefaultCloseTimeout]
//   - log: outputs certain logging from http.Server
//   - —
//   - a connection/thread may be held for long time by http/1.1 keep-alive,
//     http/2 multiplexing or connection hijacking by Websockets and similar
//   - therefore, any parallelism limit should be rather high.
//     If the total number of process goroutines is greater than 1,024,
//     the Go runtime temporary memory leaks from internal slices and maps
//     becomes significant
//   - CDavServer is always heap allocated due to http.Server code
//   - [http.Server] logs:
//   - — errors accepting connections
//   - — unexpected behavior from handlers
//   - — underlying FileSystem errors
//   - — there is http.logf: error when listing directory, bad response.WriteHeader call
//   - — http.Server.logf: bad WriteHeader calls, bad Content-Length, panic in Mux,
//     TLS handshake error, Accept error
func NewServer(
	parallelism int,
	ctx context.Context,
	closeTimeout time.Duration,
	log parl.PrintfFunc,
	serverStruct *http.Server,
) (c *Server) {
	parl.NilPanic("serverStruct", serverStruct)

	c = &Server{
		httpServer:   http.Server{},
		closeTimeout: closeTimeout,
		log:          log,
		ctx:          ctx,
	}
	phlib.CopyServer(serverStruct, &c.httpServer)
	if c.httpServer.ErrorLog == nil {
		c.httpServer.ErrorLog = NewErrorLog(log)
	}
	if parallelism != 0 {
		c.threadLimiter = parl.NewModeratorCore(parallelism)
	}
	if c.httpServer.BaseContext == nil {
		c.httpServer.BaseContext = c.getBaseContext
	}
	return
}

// Listen
//   - addr: listen address like “1.2.3.4:443” or domain name “localhost:1234”.
//     Host and port with colon separator or colon-port must be specified.
//     Empty host means listen to all interfaces.
//     Port zero means select a high port.
//   - addr [DefaultAddr]: “:1034” ie. listening on all interfaces
//   - —
//   - net.Listener interface, runtime-type is package-private tls
func (c *Server) Listen(binaryCert parl.CertificateDer, privateKey crypto.Signer, addr ...pnet.SocketAddress) (tlsListener net.Listener, near netip.AddrPort, err error) {

	// check for closed server
	if c.inShutdown.Load() {
		err = perrors.ErrorfPF("%w", http.ErrServerClosed)
		return
	}

	// get socket address where the server should listen
	var socketAddress pnet.SocketAddress
	if len(addr) > 0 {
		socketAddress = addr[0]
	} else {
		socketAddress = pnet.NewSocketAddressFromString(pnet.NetworkTCP, c.httpServer.Addr)
	}

	// create TLS listener for https
	if tlsListener, err = ListenTLS(socketAddress, &c.httpServer, binaryCert, privateKey); err != nil {
		return
	}

	// tlsListener *tls.listener
	//	- tlsListener is package-private TLS implementation
	// parl.D("tlsListener %T", tlsListener)

	// get resulting near socket-address: runtime-type: *net.TCPAddr
	if near, err = pnet.AddrPortFromAddr(tlsListener.Addr()); err != nil {
		// unparseable near address: should never happen
		_ /*err*/ = tlsListener.Close()
	}

	return
}

// Serve delegates to [http.Server.Serve] with optionally a server-wide
// thread-limiting listener returned by [Server.Listen]
//   - the server may operate any number of listeners
func (c *Server) Serve(listener net.Listener) (err error) {

	// check for closed server
	if c.inShutdown.Load() {
		err = perrors.ErrorfPF("%w", http.ErrServerClosed)
		return
	}

	// implement optional limiting of parallelism
	if c.threadLimiter != nil {
		var listenerOnce = MakeListenerOnce(listener, c.threadLimiter)
		listener = &listenerOnce
	}

	// Serve listens to listener until Accept returns error
	//	- if Close or Shutdown was invoked,
	//		ErrServerClosed is returned prior to Accept
	if err = c.httpServer.Serve(listener); errors.Is(err, http.ErrServerClosed) {
		// Accept returned after invocation of Close or Shutdown

		// err: *errors.errorString "http: Server closed"
		// parl.D("err: %T %[1]q", err)

		err = nil
	}

	return
}

// RegisterOnShutdown thread-safe
func (c *Server) RegisterOnShutdown(f func()) { c.httpServer.RegisterOnShutdown(f) }

// Close is graceful Close await request completion for default 3 seconds
//   - similar to [http.Server.Shutdown]
func (c *Server) Close() (err error) {
	if !c.inShutdown.Load() {
		c.inShutdown.Store(true)
	}

	// try 3-second graceful shutdown
	var shutdownCtx, cancelFunc = context.WithTimeout(context.Background(), c.closeTimeout)
	defer cancelFunc()

	// attempt to shut down server gracefully
	if e := c.httpServer.Shutdown(shutdownCtx); e == nil {
		// good graceful shutdown return
		return
	} else if e != context.DeadlineExceeded && e != context.Canceled {
		err = perrors.ErrorfPF("http.Shutdown %w", e)
		return // bad graceful shutdown return
	}
	// graceful shutdown timed out

	return c.CloseNow()
}

// CloseNow shuts down tearing down any active connections
//   - similar to [http.Server.Close]
func (c *Server) CloseNow() (err error) {
	if !c.inShutdown.Load() {
		c.inShutdown.Store(true)
	}
	if err = c.httpServer.Close(); err != nil {
		err = perrors.ErrorfPF("http Close %w", err)
	}
	return
}

// getBaseContext is implementation for [http.Server.BaseContext]
func (c *Server) getBaseContext(listener net.Listener) (ctx context.Context) { return c.ctx }

// - func (s *http.Server) Serve(l net.Listener) error
// - drawback: one goroutine for each connection
var _ = (&http.Server{}).Serve

// Shutdown
//   - func (s *http.Server) Shutdown(ctx context.Context) error
var _ = (&http.Server{}).Shutdown

// Request handling
//   - eventually http requests are handled by a handler, a method with signature of [http.Handler]
//   - ServeHTTP(ResponseWriter, *Request)
//   - Request is struct with Method, URL, Header and Body fields
//   - ResponseWriter is interface with Header() Write() WriteHeader()
var _ http.Handler

// ResponseWriter interface
var _ http.ResponseWriter

// Request struct
var _ http.Request

// - [http.ListenAndServe] is a simple function delegating to [http.Server.ListenAndServe]
var _ = http.ListenAndServe

// [http.HandleFunc] is handler signature for top-level function
var _ = http.HandleFunc

// ServeMux directs incoming requests to handlers based on URL path
var _ http.ServeMux

// [http.Server.ListenAndServe] delegates to [net.Listen] then
// a loop with [net.Listen.Accept] returning incoming connections
//   - eventually will return error eg. ErrServerClosed
//   - func (s *http.Server) ListenAndServe() error
var _ = (&http.Server{}).ListenAndServe
