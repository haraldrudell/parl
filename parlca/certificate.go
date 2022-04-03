/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/ed25519"
	"crypto/x509"

	"github.com/haraldrudell/parl/perrors"
)

// x509Certificate extends x509.Certificate with utility methods
type x509Certificate struct {
	*x509.Certificate
}

func (c *x509Certificate) IsValid() (isValid bool) {
	if !c.HasPublic() {
		return
	}
	cert := c.Certificate
	if cert.SerialNumber == nil ||
		cert.Issuer.CommonName == "" ||
		len(cert.Issuer.Country) == 0 ||
		cert.NotBefore.IsZero() ||
		cert.NotAfter.IsZero() ||
		cert.KeyUsage == 0 {
		return
	}
	isValid = true
	return
}

func (c *x509Certificate) HasPublic() (hasPublic bool) {
	if len(c.PublicKeyBytes()) == 0 ||
		c.Certificate.PublicKeyAlgorithm == x509.UnknownPublicKeyAlgorithm {
		return
	}
	hasPublic = true
	return
}

func (c *x509Certificate) PublicKeyBytes() (bytes []byte) {
	if c == nil {
		return
	}
	cert := c.Certificate
	if cert == nil {
		return
	}
	ed25519PublicKey, ok := cert.PublicKey.(ed25519.PublicKey)
	if !ok {
		panic(perrors.Errorf("Bad PublicKey type: %T", cert.PublicKey))
	}
	bytes = ed25519PublicKey
	return
}
