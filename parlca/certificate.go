/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/x509"
	"encoding/pem"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// Certificate wraps a der format x509 certificate.
//   - der-format certificate is produced by [x509.CreateCertificate]
//   - An [x509.Certificate] can be obtained from [x509.ParseCertificate]
type Certificate struct {
	/*
		// x509.Certificate

		CheckCRLSignature(crl *pkix.CertificateList) (err error)
		CheckSignature(algo x509.SignatureAlgorithm, signed, signature []byte) (err error)
		CheckSignatureFrom(parent *x509.Certificate) (err error)
		CreateCRL(rand io.Reader, priv any, revokedCerts []pkix.RevokedCertificate,
			now, expiry time.Time) (crlBytes []byte, err error)
		Equal(other *x509.Certificate) (isEqual bool)
		Verify(opts x509.VerifyOptions) (chains [][]*x509.Certificate, err error)
		VerifyHostname(host string) (err error)
	*/

	der parl.CertificateDer
}

// NewCertificate returns an object that can produce:
//   - textual pem format and
//   - expanded [x509.Certificate] format
//   - storage is maximum efficient binary der asn.1 format
//   - [Certificate.DER] binary data
//   - [Certificate.PEM] textual block
//   - [Certificate.ParseCertificate] [x509.Certificate] data structure
func NewCertificate(certificateDer parl.CertificateDer) (certificate parl.Certificate) {
	return &Certificate{der: certificateDer}
}

// 221121 don’t know what this is. Make it compile
func LoadCertificate(filename string) {}

/*
	func (c *Certificate) IsValid() (isValid bool) {
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

	func (c *Certificate) HasPublic() (hasPublic bool) {
		if len(c.PublicKeyBytes()) == 0 ||
			c.Certificate.PublicKeyAlgorithm == x509.UnknownPublicKeyAlgorithm {
			return
		}
		hasPublic = true
		return
	}

	func (c *Certificate) PublicKeyBytes() (bytes []byte) {
		if c == nil {
			return
		}
		cert := c.Certificate
		if cert == nil {
			return
		}
		//ed25519PublicKey, ok := cert.PublicKey.(*rsa.PublicKey)
		ok := false

		//ed25519PublicKey, ok := cert.PublicKey.(ed25519.PublicKey)
		if !ok {
			panic(perrors.Errorf("Bad PublicKey type: %T", cert.PublicKey))
		}
		//bytes = ed25519PublicKey
		return
	}
*/

// DER returns the binary der asn.1 format of the certificate
func (c *Certificate) DER() (certificateDer parl.CertificateDer) { return c.der }

// PEM returns a file-writable and human-readable pem block
//   - “==… CERTIFICATE…”
func (c *Certificate) PEM() (pemBytes parl.PemBytes) {
	return append(
		// lead-in text for pem block sha256 and sha1 fingerprint
		[]byte(PemText(c.der, c.der)),
		// ==… CERTIFICATE…
		pem.EncodeToMemory(&pem.Block{
			Type:  pemCertificateType,
			Bytes: c.DER(),
		})...)
}

// ParseCertificate returns expanded [x509.Certificate] format
//   - allows certificate to be used as parent argument to [x509.CreateCertificate]
//   - provides access to certificate datapoints
func (c *Certificate) ParseCertificate() (certificate *x509.Certificate, err error) {
	if certificateDer := c.der; len(certificateDer) == 0 {
		err = perrors.New("certificate der uninitialized")
		return
	} else if certificate, err = x509.ParseCertificate(certificateDer); err != nil {
		err = perrors.ErrorfPF("x509.ParseCertificate: %w", err)
	}

	return
}
