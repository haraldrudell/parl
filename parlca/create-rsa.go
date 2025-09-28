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

// CreateRSA creates server credentials for testing without browser
// in 262 ms.
// RSA credentials can be parsed from pem in 850 µs.
// Ed25519 credential can be generated in 2 ms.
// calib exports rsa-2048 localhost credentials
//   - to get x509 struct use cert.ParseCertificate
//   - consider using [github.com/haraldrudell/parl/parlca/calib.CertLocalhost] RSA
//   - certificate canonical name is host’s name, valid for 10 years, 2,048 bits
//   - panic on error which never happens
func CreateRSA() (cert parl.Certificate, key parl.PrivateKey) {
	var err error
	cert, key, _ /*caCertDER*/, _ /*caKey*/, err = CreateCredentials(x509.RSA, DefaultCAName)
	if err != nil {
		err = perrors.ErrorfPF("CreateCredentials %w", err)
		panic(err)
	}

	return
}
