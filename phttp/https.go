/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phttp

import (
	"context"
	"crypto"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pnet"
)

const (
	// default port for https: 443, or address for localhost IPv4 or IPv6 port 443
	HttpsPort uint16 = 443
)

// Https is an https server instance encapsulating its credentials
//   - [Https] is an [http.Server] that:
//   - — uses in-memory credentials
//   - — has route-handler configuration
//   - — provides awaitable threading for any number of listeners
//   - — has awaitable first error and shutdown
//   - — provides streaming or collectable [net.Listener.Accept] errors
type Https struct {
	// Server holds handler mux and fields like TLSConfig
	http.Server
	// the request multiplexer as instance
	//	- http.Server.Handler is interface version
	serveMux *http.ServeMux
	// server certificate in binary DER ASN.1 format
	cert parl.CertificateDer
	// private key in binary format with signing methods
	private crypto.Signer
	// wg awaits:
	//	- all [Https.Serve] threads if any exiting and
	//	- first [Https.Shutdown] invocation complete
	wg parl.WaitGroupCh
	// shutdownOnce indicates that [Https.Shutdown] has been invoked
	shutdownOnce parl.OnceCh
	// errs collect errors from exiting [Https.Serve] threads
	errs parl.ErrSlice
}

// NewHttps creates https server with any number of listeners
//   - [Https] is an [http.Server] that:
//   - — uses in-memory credentials
//   - — has route-handler configuration
//   - — provides awaitable threading for any number of listeners
//   - — has awaitable first error and deferrable shutdown
//   - — provides streaming or collectable [net.Listener.Accept] errors
//   - certDER: certificate in binary DER ASN.1 format.
//     Can be obtained from [parl.CertificateAuthority.Sign]
//   - signer: private key material and signing methods.
//     Can be obatined from [parlca.NewPrivateKey]
//   - log: optional function receiving log output
//   - — if missing, output may occur on standard error
//   - [Https.HandleFunc] configures handle functions
//   - [Https.Listen] creates TLS listener
//   - [Https.Serve] gorouting invoking [net.Listener.Accept]
//   - [Https.Errs] returns error iterator
//   - [Https.ShutdownCh] awaits Shutdown complete
//   - [Https.Shutdown] graceful shutdown
//   - [Https.Shutdown2] deferrable shutdown
//   - error sources:
//   - — returned by [Https.Listen]: troubles in listener set-up
//   - — returned by [Https.Shutdown] [Https.Shutdown2] listener shutdown errors
//   - — [Https.Errs] real-time streaming iterator erorrs or error collection after shutdown
//   - awaitable via [Https.ShutdownCh] that awaits completion of Shutdown invocaion
//     and exit of all Serve threads.
//     Serve threads that do not exit causes wait to be indefinite
//
// Usage:
//
//	var s = NewHttps(cert, key)
//	defer s.Shutdown2(&err)
//	s.HandleFunc("/", myHandler)
//	if nearAddrPort, listener, e := s.Listen(socketAddress); e == nil {
//	  go s.Serve(listener)
//	  println(nearAddrPort.String())
//	} else {
//	  err = e
//	  return
//	}
//	<-s.ShutdownCh()
//	parl.DeferredErrorSource(s.Errs(), &err)
//	return
func NewHttps(certDER parl.CertificateDer, signer crypto.Signer, log ...parl.PrintfFunc) (httpsServer *Https) {
	httpsServer = &Https{
		serveMux: http.NewServeMux(),
		cert:     certDER,
		private:  signer,
	}
	httpsServer.wg.Add(1)
	// request multiplexer as interface
	httpsServer.Server.Handler = httpsServer.serveMux
	if len(log) > 0 {
		if f := log[0]; f != nil {
			httpsServer.Server.ErrorLog = NewErrorLog(f)
		}
	}
	return
}

// Errs returns a streaming error iterator or post serve-exit error collection
func (s *Https) Errs() (errIterator parl.ErrsIter) { return &s.errs }

// ShutdownCh triggers when the first Shutdown has completes
func (s *Https) ShutdownCh() (ch parl.AwaitableCh) { return s.wg.Ch() }

// HandlerFunc is the signature for URL handlers
type HandlerFunc func(http.ResponseWriter, *http.Request)

// HandleFunc registers a URL-handler for the server
//   - pattern described by [http.ServeMux]: [METHOD ][HOST]/[PATH]
//   - — pattern empty string or invalid is panic
//   - — subsequent identical pattern is panic
//   - handler is func(http.ResponseWriter, *http.Request)
//   - — handler nil is panic
//   - — panic in handler is logged by http package
func (s *Https) HandleFunc(pattern string, handler HandlerFunc) {
	s.serveMux.HandleFunc(pattern, handler)
}

// Listen initiates listening and returns the error channel
//   - can only be invoked once or panic
//   - errCh closes on server shutdown
//   - non-blocking, all errors are sent on the error channel
//   - if nearSocket.Addr is invalid, all interfaces for IPv6 if allowed, IPv4 otherwise is used
//   - if nearSocket.Port is zero:
//   - — if network is NetworkDefault: ephemeral port
//   - — otherwise port 443 “:https” is used
//   - for NetworkDefault, NetworkTCP is used
//   - panic for bad Network
func (s *Https) Listen(socketAddress pnet.SocketAddress) (
	nearAddrPort netip.AddrPort,
	listener net.Listener,
	err error,
) {
	// get near tls socket listener
	// *tls.listener — need this or it’s file certificates
	var tlsListener net.Listener
	if tlsListener, err = s.TLS(socketAddress); err != nil {
		return
	}
	defer s.maybeClose(&tlsListener, &err)

	// set nearAddrPort
	if nearAddrPort, err = pnet.AddrPortFromAddr(tlsListener.Addr()); err != nil {
		return
	}
	listener = tlsListener
	tlsListener = nil

	return
}

// TLS creates a TLS listener
//   - can only be invoked once or panic
//   - errCh closes on server shutdown
//   - non-blocking, all errors are sent on the error channel
//   - if nearSocket.Addr is invalid, all interfaces for IPv6 if allowed, IPv4 otherwise is used
//   - if nearSocket.Port is zero:
//   - — if network is NetworkDefault: ephemeral port
//   - — otherwise port 443 “:https” is used
func (s *Https) TLS(socketAddress pnet.SocketAddress) (tlsListener net.Listener, err error) {
	tlsListener, err = ListenTLS(
		socketAddress,
		&s.Server,
		s.cert,
		s.private,
	)

	return
}

func (s *Https) GetServeGoFunction() (serve func(listener net.Listener)) {
	s.wg.Add(1)
	serve = s.serve
	return
}

// Serve is a thread invoking [http.Server.Serve] providing:
//   - await via wg
//   - error collection
func (s *Https) serve(listener net.Listener) {
	var err error
	// invoked without recover
	defer s.closeErrsOn1()
	defer s.wg.Done()
	defer parl.Recover2(func() parl.DA { return parl.A() }, &err, &s.errs)

	// Serve is blocking invocations of [net.ListenerAccept]
	//	- Serve will always return non-nil error
	err = s.Server.Serve(listener)

	// examine error
	if s.shutdownOnce.IsInvoked() && errors.Is(err, http.ErrServerClosed) {
		// Shutdown was invoked and
		//	- error is http.ErrServerClosed
		//	- ignore that error, it is gracefule xit
		err = nil
	}
}

// Shutdown shuts down the server gracefully
//   - ctx: optional context for how long to wait for threads to exit
//   - — default: 3 s
//   - err: only the fiurst invocation receive errors
func (s *Https) Shutdown(ctx ...context.Context) (err error) {

	// first thread closes server, subsequent threads wait for close
	if isWinner, done := s.shutdownOnce.IsWinner(); !isWinner {
		return
	} else {
		defer s.wg.Done()
		defer s.closeErrsOn1()
		defer done.Done()
	}

	// context with max wait time for request completion
	var shutdownCtx context.Context
	if len(ctx) > 0 {
		shutdownCtx = ctx[0]
	}
	if shutdownCtx == nil {
		var cancelFunc context.CancelFunc
		// try 3-second graceful shutdown
		shutdownCtx, cancelFunc = context.WithTimeout(context.Background(), httpShutdownTimeout)
		defer cancelFunc()
	}

	// attempt to shut down server gracefully
	if err = s.Server.Shutdown(shutdownCtx); err == nil {
		return // good graceful shutdown return
	} else if err != context.DeadlineExceeded && err != context.Canceled {
		err = perrors.ErrorfPF("http.Shutdown %w", err)
		return // bad graceful shutdown return
	}

	// forced shut down
	if err = s.Close(); err != nil {
		err = perrors.ErrorfPF("http.Close %w", err)
	}

	return
}

// Shutdown2 is deferrable Shutdown
func (s *Https) Shutdown2(errp *error, ctx ...context.Context) {
	var err error
	if err = s.Shutdown(ctx...); err == nil {
		return
	}
	*errp = perrors.AppendError(*errp, err)
}

// closes listener if non-nil
func (s *Https) maybeClose(listenerp *net.Listener, errp *error) {
	var listener = *listenerp
	if listener == nil {
		return
	}
	parl.Close(listener, errp)
}

// closeErrsOn1 ensures that shutdown is invoked
//   - invoked on:
//   - — all Serve-thread exits and
//   - — completion of winner Shutdown
//     -
//   - closes [Https.Errs]
func (s *Https) closeErrsOn1() {

	// if shutdown has not been invoked so:
	//	- this is the first exiting Serve thread
	//	- must invoke Shutdown to cause other Serve threads to exit
	//	- upon return from Shutdown, it is certain that the server is shut down
	//	- the error slice can be closed
	if !s.shutdownOnce.IsInvoked() {
		if err := s.Shutdown(); err != nil {
			s.errs.AddError(err)
		}
		// after shutdown completion,
		// it is likely that allServe threads have exited
		//	- close error slice
		s.errs.EndErrors()
		return
	}

	// Shutdown has been invoked so this is:
	//	- an effective shutdown invocation prior to any Serve thread exit
	//	- — one currentCount is held by Shutdown
	//	- a Serve thread exiting upon Shutdown invocation
	//	- — Serve thread has already invoked wg.Done
	//	- error slice should be closed when:
	//	- — there are no Serve threads: totalAdds == 1
	//	- — this is the last exiting Serve thread: currentCount == 1
	if currentCount, totalAdds := s.wg.Counts(); totalAdds == 1 || currentCount == 1 {
		s.errs.EndErrors()
	}
}

const (
	http11              = "http/1.1"
	httpShutdownTimeout = 5 * time.Second
)
