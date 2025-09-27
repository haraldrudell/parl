/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phttp

import (
	"crypto"
	"crypto/tls"
	"io/fs"
	"net"
	"net/http"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pnet"
	"github.com/haraldrudell/parl/pstrings"
)

// ListenTLS returns a listening near TLS socket enabling use of the https protocol
// with [http.Server]
//   - allows using binary credentials from memory
//   - socketAddress: a container that can hold either an IP literal or
//     a domain name. Returned by:
//   - — [pnet.NewSocketAddressLiteral]
//   - — [pnet.NewSocketAddress]
//   - httpServer: pointer to the http server’s populated struct
//   - binaryCert: server certificate in binary form
//   - — from [parlca.NewCertificate]
//   - — from [parlca.ReadPemFromFile]
//   - — from [x509.Certificate.Raw]
//   - privateKey: server’s private key in binary form.
//   - — implementation depends on signing algorithm
//   - — from [parlca2.Credentials]
//   - — from [parlca.ReadPemFromFile]
//   - tlsListener: configured TLS listener.
//   - — value is [net.Listener]: Accept/Close/Addr
//   - — runtime value is package-private value
//     returned by [tls.NewListener]
//   - err: listen to socketAddress failed
func ListenTLS(
	socketAddress pnet.SocketAddress,
	httpServer *http.Server,
	binaryCert parl.CertificateDer,
	privateKey crypto.Signer,
) (tlsListener net.Listener, err error) {

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
	if listener, err = pnet.Listen(socketAddress); err != nil {
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
	var tlsCertificate = &tlsConfig.Certificates[0]
	// certificate not from file system
	tlsCertificate.Certificate = append(tlsCertificate.Certificate, binaryCert)
	// private key not from file system
	tlsCertificate.PrivateKey = privateKey
	tlsListener = tls.NewListener(listener, tlsConfig)

	return

}
