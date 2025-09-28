/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"crypto/tls"
	"crypto/x509"
)

// NewTLSConfig returns a TLS configuration that has
// cert as the only root certificate
//   - self-signed server-certificates needs this for http client to access a server
func NewTLSConfig(cert *x509.Certificate) (tlsConfig *tls.Config) {
	var certPool = x509.NewCertPool()
	certPool.AddCert(cert)
	tlsConfig = &tls.Config{RootCAs: certPool}

	return
}
