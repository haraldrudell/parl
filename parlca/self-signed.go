/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

const (
	/*
		NoPassword       PasswordType = "\tnoPassword"
		GeneratePassword PasswordType = "\tgeneratePassword"
		GenerateOnTheFly Strategy     = iota << 0
		UseFileSystem
		DefaultStrategy = GenerateOnTheFly
	*/
	DefaultCountry  = "US" // certificate country: US
	notAfterYears   = 10   // certificate validity for 10 years
	caSubjectSuffix = "ca" // ca appended to commonName
)

type SelfSigned struct {
	parl.Certificate // DER() PEM()
	PrivateKey       parl.PrivateKey
}

// NewSelfSigned creates a self-ssigned certificate authority
//   - ca: a certificate with embedded private key
//   - — implementation is parlca.Certificate, ie. binary der format
//   - canonicalName: 240623 not used
//   - algo:
//   - — [x509.Ed25519] smallest key size but as of 2024 not supported by browsers
//   - — [x509.RSA] the most commonly used algorithm for browsers
//   - — [x509.ECDSA]
func NewSelfSigned(canonicalName string, algo x509.PublicKeyAlgorithm) (ca parl.CertificateAuthority, err error) {
	c := SelfSigned{}

	// create private key
	if c.PrivateKey, err = NewPrivateKey(algo); err != nil {
		return
	}

	// create certificate of certificate authority
	var certificateDer parl.CertificateDer
	var cert = &x509.Certificate{}
	EnsureSelfSigned(cert)
	if certificateDer, err = x509.CreateCertificate(rand.Reader,
		cert,                  // template
		cert,                  // parent
		c.PrivateKey.Public(), // pub any: *rsa.PublicKey *ecdsa.PublicKey ed25519.PublicKey
		c.PrivateKey,          // priv any: crypto.Signer
	); perrors.IsPF(&err, "x509.CreateCertificate %w", err) {
		return
	}
	c.Certificate = NewCertificate(certificateDer)
	ca = &c
	return
}

func NewSelfSigned2(privateKey parl.PrivateKey, certificate parl.Certificate) (ca parl.CertificateAuthority) {
	return &SelfSigned{Certificate: certificate, PrivateKey: privateKey}
}

func (ca *SelfSigned) Sign(template *x509.Certificate, publicKey crypto.PublicKey) (certDER parl.CertificateDer, err error) {

	// get certificate authority x509.Certificate
	var caCert *x509.Certificate
	if caCert, err = ca.Check(); err != nil {
		return
	}
	caSigner := ca.PrivateKey //.Private()

	// sign template
	if certDER, err = x509.CreateCertificate(rand.Reader, template, caCert, publicKey, caSigner); err != nil {
		err = perrors.Errorf("x509.CreateCertificate: '%w'", err)
		return
	}
	return
}

func (ca *SelfSigned) Check() (cert *x509.Certificate, err error) {
	if err = ca.PrivateKey.Validate(); err != nil {
		return
	} else if cert, err = ca.ParseCertificate(); perrors.IsPF(&err, "x509.ParseCertificate: '%w'", err) {
		return
	} else if cert.PublicKey == nil {
		err = perrors.NewPF("public key uninitialized")
		return
	}
	return
}

func (ca *SelfSigned) Private() (privateKey parl.PrivateKey) { return ca.PrivateKey }
