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
// cert as a root certificate
//   - this allows a client to access a server that has self-signed certificate
func NewTLSConfig(cert *x509.Certificate) (tlsConfig *tls.Config) {
	certPool := x509.NewCertPool()
	certPool.AddCert(cert)
	return &tls.Config{RootCAs: certPool}
}
