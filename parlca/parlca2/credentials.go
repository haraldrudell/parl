/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca2

import (
	"crypto"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"net"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/parlca"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pnet"
)

// Credentials returns minimum credentials for TLS
//   - x509Certificate: IP-literal based certificate for “127.0.0.1” and “::1’
//   - — x509Certificate.Raw is binary certificate DER ASN.1
//   - — signed by the private key
//   - the private key is [crypto.Signer]: Public Sign methods
func Credentials() (x509Certificate *x509.Certificate, privateKey crypto.Signer, err error) {

	// certificate output format conclusion: *x509.Certificate
	//	- [x509.CertPool.AddCert] requires [x509.Certificate]
	//	- [tls.NewListener] requires [tls.Config.Certificates] which is
	//		[tls.Certificate] which contains binary certificate DER
	//	- — having the output type being [tls.Certificate] costs one allocation
	//	- — [tls.Config.Leaf] is parsed [x509.Certificate]
	//	- output from [x509.CreateCertificate] is binary certificate DER
	var _ *tls.Certificate
	var _ *x509.Certificate

	// private key output conslusion: crypto.Signer interface,
	// parl type value
	//	- best to use [crypto.Signer] that has Public and Sign
	//	- legacy [crypto.PrivateKey] is any
	//	- using [parl.PrivateKey]
	//	- parlca types like [parlca.Ed25519PrivateKey] adds methods
	//		like Algo at no storage cost
	var _ crypto.Signer
	var _ crypto.PrivateKey

	// create private key
	if privateKey, err = parlca.MakeEd25519(); err != nil {
		return
	}

	// create certificate binary DER ASN.1
	//	- certificate for localhost “::1” “127.0.0.1”
	var template = x509.Certificate{
		IPAddresses: []net.IP{pnet.IPv4loopback, net.IPv6loopback},
		DNSNames:    []string{pnet.LocalHost},
	}
	parlca.EnsureTemplate(&template)
	// certificate use is server authentication
	parlca.EnsureServer(&template)
	var certificateDer parl.CertificateDer
	if certificateDer, err = x509.CreateCertificate(rand.Reader,
		&template,           // template
		&template,           // parent
		privateKey.Public(), // pub any: *rsa.PublicKey *ecdsa.PublicKey ed25519.PublicKey
		privateKey,          // priv any: crypto.Signer
	); perrors.IsPF(&err, "x509.CreateCertificate %w", err) {
		return
	}

	// parse into x509.Certificate
	var certificate = parlca.NewCertificate(certificateDer)
	x509Certificate, err = certificate.ParseCertificate()

	return
}
