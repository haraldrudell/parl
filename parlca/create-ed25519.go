/*
© 2023–present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parlca

import (
	"crypto/x509"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// CreateEd25519 creates server credentials for testing without browser
// in less than 2 ms
//   - macOS 26.0 250927 if certificate’s public key algorithm is ed25519,
//     TLS handshake error is “certificate is using a broken key size”
//   - — “certificate is not standards compliant” means
//     unknown self-signed certificate authority
//   - consider using [github.com/haraldrudell/parl/parlca/calib.CertLocalhost] RSA
//   - to get x509 struct use cert.ParseCertificate
//   - certificate cacnonical name is host’s name, valid for 10 years
//   - panic on error which never happens
func CreateEd25519() (cert parl.Certificate, key parl.PrivateKey) {
	var err error
	cert, key, _ /*caCertDER*/, _ /*caKey*/, err = CreateCredentials(x509.Ed25519, DefaultCAName)
	if err != nil {
		err = perrors.ErrorfPF("CreateCredentials %w", err)
		panic(err)
	}

	return
}
