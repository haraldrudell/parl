/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phttp

import (
	"context"
	"crypto"
	"crypto/tls"
	"errors"
	"io/fs"
	"net"
	"net/http"
	"net/netip"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pnet"
	"github.com/haraldrudell/parl/pstrings"
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
	// errs collect errors from exing [Https.Serve] threads
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
			httpsServer.Server.ErrorLog = NewLogCapturer(f)
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
	var httpServer = &s.Server

	// make ServeTLS execute srv.setupHTTP2_ServeTLS go 1.16.6
	var badPath = "%"
	if err = httpServer.ServeTLS(nil, badPath, badPath); err == nil {
		err = perrors.New("missing error from srv.ServeTLS")
		return
	}

	// ignore the error if expected badPath
	if pathError, ok := err.(*fs.PathError); ok {
		if pathError.Path == badPath {
			err = nil
		}
	}

	// failure on other errors
	if err != nil {
		err = perrors.Errorf("srv.ServeTLS: '%w'", err)
		return
	}

	// get *net.TCPListener from srv.ListenAndServeTLS

	// underlying tcp listener from pnet
	var listener net.Listener
	if listener, err = pnet.Listen(socketAddress, pnet.NoCancel); err != nil {
		return
	}
	var _ = net.Listen

	// create a TLS listener from srv.ServeTLS
	var tlsConfig *tls.Config
	if httpServer.TLSConfig == nil {
		tlsConfig = &tls.Config{}
	} else {
		tlsConfig = httpServer.TLSConfig.Clone()
	}
	if !pstrings.StrSliceContains(tlsConfig.NextProtos, http11) {
		tlsConfig.NextProtos = append(tlsConfig.NextProtos, http11)
	}
	tlsConfig.Certificates = make([]tls.Certificate, 1)
	tlsCertificate := &tlsConfig.Certificates[0]
	tlsCertificate.Certificate = append(tlsCertificate.Certificate, s.cert) // certificate not from file system
	tlsCertificate.PrivateKey = s.private                                   // private key not from file system
	tlsListener = tls.NewListener(listener, tlsConfig)

	return
}

// Serve is a thread invoking [http.Server.Serve] providing:
//   - await via wg
//   - error collection
func (s *Https) Serve(listener net.Listener) {
	var err error
	defer s.closeErrsOn1()
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
	if isWinner, done := s.shutdownOnce.IsWinner(); !isWinner {
		return
	} else {
		defer s.wg.Done()
		defer s.closeErrsOn1()
		defer done.Done()
	}
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
	if err = s.Server.Shutdown(shutdownCtx); err == nil {
		return // good graceful shutdown return
	} else if err != context.DeadlineExceeded && err != context.Canceled {
		err = perrors.ErrorfPF("http.Shutdown %w", err)
		return // bad graceful shutdown return
	}
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
//   - invoked on Serve-thread exit and Shutdown completion
//   - closes [Https.Errs]
func (s *Https) closeErrsOn1() {

	// if shutdown has not been invoked,
	// this is the first exiting Serve thread
	//	- Shutdown should complete
	if !s.shutdownOnce.IsInvoked() {
		if err := s.Shutdown(); err != nil {
			s.errs.AddError(err)
		}
		// after shutdown completion,
		// it is likely that allServe threads have exited
		s.errs.EndErrors()
		return
	}

	// if there are no serve threads, shutdown should invoke EndErrors
	if _, totalAdds := s.wg.Counts(); totalAdds == 0 {
		s.errs.EndErrors()
	}
}

const (
	httpsAddr           = ":https"
	http11              = "http/1.1"
	httpShutdownTimeout = 5 * time.Second
)
