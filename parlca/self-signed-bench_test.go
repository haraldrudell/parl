/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/x509"
	"net"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/parlca/calib"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pnet"
)

// certificate authority and certificate-private key is created in 0.12 s
//
// go test -benchmem -run=^$ -bench ^BenchmarkSelfSigned$ github.com/haraldrudell/parl/parlca
// 240623 c66
// pkg: github.com/haraldrudell/parl/parlca
// BenchmarkSelfSigned-10    	      19	 122106472 ns/op	 2975986 B/op	    8514 allocs/op
func BenchmarkSelfSigned(b *testing.B) {
	const (
		canonicalName = ""
	)
	var (
		caCert       parl.CertificateAuthority
		caX509       *x509.Certificate
		serverSigner Ed25519PrivateKey
		template     x509.Certificate
		certDER      parl.CertificateDer
		err          error
	)
	for i := range b.N {
		_ = i
		template = x509.Certificate{
			IPAddresses: []net.IP{pnet.IPv4loopback, net.IPv6loopback},
		}
		calib.EnsureServer(&template)
		if caCert, err = NewSelfSigned(canonicalName, x509.RSA); err != nil {
			b.Fatalf("FAIL parlca.NewSelfSigned %s “%s”", x509.RSA, perrors.Short(err))
		} else if caX509, err = caCert.Validate(); err != nil {
			b.Fatalf("FAIL: caCert.Check: %s", perrors.Short(err))
		} else if serverSigner, err = MakeEd25519(); err != nil {
			b.Fatalf("FAIL server parlca.NewEd25519: “%q”", err)
		} else if certDER, err = caCert.Sign(&template, serverSigner.Public()); err != nil {
			b.Fatalf("FAIL signing server certificate: “%s”", err)
		}
		_ = certDER
		_ = caX509
	}
}
