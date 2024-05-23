/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phttp

import (
	"crypto"
	"crypto/tls"
	"io/fs"
	"net"
	"net/http"
	"net/netip"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pnet"
	"github.com/haraldrudell/parl/pstrings"
)

const (
	// default port for https: 443, or address for localhost IPv4 or IPv6 port 443
	HttpsPort uint16 = 443
)

// Http is an http server instance
//   - based on [http.Server]
//   - has listener thread
//   - all errors sent on channel
//   - idempotent deferrable panic-free shutdown
//   - awaitable, observable
type Https struct {
	Http
	Cert    parl.CertificateDer
	Private crypto.Signer
}

// NewHttp creates http server for default “localhost:443”
//   - if nearSocket.Addr is invalid, all interfaces for IPv6 if allowed, IPv4 otherwise is used
//   - if nearSocket.Port is zero:
//   - — if network is NetworkDefault: ephemeral port
//   - — otherwise port 443 “:https” is used
//   - for NetworkDefault, NetworkTCP is used
//   - panic for bad Network
//
// Usage:
//
//	var s = NewHttps(netip.AdddrPort{}, pnet.NetworkTCP)
//	s.HandleFunc("/", myHandler)
//	defer s.Shutdown()
//	for err := range s.Listen() {
func NewHttps(nearSocket netip.AddrPort, network pnet.Network, certDER parl.CertificateDer, signer crypto.Signer) (hp *Https) {
	if nearSocket.Port() == 0 && network != pnet.NetworkDefault {
		nearSocket = netip.AddrPortFrom(nearSocket.Addr(), HttpsPort)
	}
	return &Https{
		Http:    *NewHttp(nearSocket, network),
		Cert:    certDER,
		Private: signer,
	}
}

// Listen initiates listening and returns the error channel
//   - can only be invoked once or panic
//   - errCh closes on server shutdown
//   - non-blocking, all errors are sent on the error channel
func (s *Https) Listen() (errCh parl.Errs) {
	if !s.NoListen.CompareAndSwap(false, true) {
		panic(perrors.NewPF("multiple invocations"))
	}
	errCh = &s.ErrCh
	// listen is deferred so just launch the thread
	go s.httpsListenerThread()
	return
}

const (
	httpsAddr = ":https"
	http11    = "http/1.1"
)

// httpsListenerThread is gorouitn starting listen and
// waiting for server to terminate
func (s *Https) httpsListenerThread() {
	defer s.EndListenAwaitable.Close()
	var err error
	defer parl.Recover(func() parl.DA { return parl.A() }, &err, &s.ErrCh)

	// get near tls socket listener
	// *tls.listener — need this or it’s file certificates
	var tlsListener net.Listener
	if tlsListener, err = s.TLS(); err != nil {
		return
	}
	defer s.maybeClose(&tlsListener, &err)

	// set Near socket address
	if s.Near, err = pnet.AddrPortFromAddr(tlsListener.Addr()); err != nil {
		return
	}
	s.ListenAwaitable.Close()

	// blocks here until Shutdown or Close
	err = s.Server.Serve(tlsListener)
	tlsListener = nil

	// on regular close, http.ErrServerClosed
	if err == http.ErrServerClosed {
		err = nil // ignore error
		return    // successful return
	}
	err = perrors.Errorf("srv.ServeTLS: ‘%w’", err)
}

func (s *Https) TLS() (tlsListener net.Listener, err error) {
	var httpServer = &s.Server

	// make ServeTLS execute srv.setupHTTP2_ServeTLS go 1.16.6
	badPath := "%"
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

	// tcp listener from pnet
	var listener net.Listener
	if listener, err = pnet.Listen(s.Network, s.Server.Addr, &s.Cancel); err != nil {
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
	tlsCertificate.Certificate = append(tlsCertificate.Certificate, s.Cert) // certificate not from file system
	tlsCertificate.PrivateKey = s.Private                                   // private key not from file system
	tlsListener = tls.NewListener(listener, tlsConfig)

	return
}

// closes listener if non-nil
func (s *Https) maybeClose(listenerp *net.Listener, errp *error) {
	var listener = *listenerp
	if listener == nil {
		return
	}
	parl.Close(listener, errp)
}
