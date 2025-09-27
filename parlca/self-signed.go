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
	DefaultCAName = ""
)

// SelfSigned is self0signed certificate authority
type SelfSigned struct {
	// DER() PEM()
	parl.Certificate
	PrivateKey parl.PrivateKey
}

// SelfSigned is [parl.CertificateAuthority]
var _ parl.CertificateAuthority = &SelfSigned{}

// NewSelfSigned creates a 10-year self-signed certificate authority
//   - canonicalName: a name that uniquely identifies the CA.
//     Signed certificates will refer to this string to identify the authority.
//   - canonicalName [DefaultCAName] "": use default
//   - — default common-name is “[hostname]ca-241231”
//     ie. the host’s name, the word ca and a date6
//   - algo:
//   - — [x509.Ed25519] smallest key size but as of 2024 not supported by browsers
//   - — [x509.RSA] the most commonly used algorithm for browsers
//   - — [x509.ECDSA]
//   - ca: certificate authority with embedded private key
//   - err: key generation failure, certificate creation failure
//   - —
//   - because [parl.CertificateAuthority] requires pointer, a pointer is returned
//   - fieldp allows for stack-allocated or use of field from previous allocation
//   - — implementation is [parlca.Certificate], ie. binary der format
func NewSelfSigned(
	canonicalName string,
	algo x509.PublicKeyAlgorithm,
	fieldp ...*SelfSigned,
) (ca *SelfSigned, err error) {

	// get ca
	if len(fieldp) > 0 {
		ca = fieldp[0]
	}
	if ca == nil {
		ca = &SelfSigned{}
	}

	// create private key
	//	- some algroithms can only generate pointer, so use pointer
	if ca.PrivateKey, err = NewPrivateKey(algo); err != nil {
		return
	}

	// create certificate of certificate authority

	// der is binary returned by x509 package
	var certificateDer parl.CertificateDer
	// cert is temporary x509 struct to create a binary byte-slice
	var cert = x509.Certificate{}
	cert.Issuer.CommonName = canonicalName
	EnsureSelfSigned(&cert)
	if certificateDer, err = x509.CreateCertificate(
		rand.Reader,
		// template
		&cert,
		// parent
		&cert,
		// pub any: *rsa.PublicKey *ecdsa.PublicKey ed25519.PublicKey
		ca.PrivateKey.Public(),
		// priv any: [crypto.Signer]
		ca.PrivateKey,
	); perrors.IsPF(&err, "x509.CreateCertificate %w", err) {
		return
	}
	ca.Certificate = NewCertificate(certificateDer)

	return
}

// NewSelfSigned creates a self-signed certificate authority
// from existing cert and key possibly read from storage
func NewSelfSignedFromKeyCert(
	privateKey parl.PrivateKey,
	certificate parl.Certificate,
	fieldp ...*SelfSigned,
) (ca *SelfSigned) {

	// get ca
	if len(fieldp) > 0 {
		ca = fieldp[0]
	}
	if ca == nil {
		ca = &SelfSigned{}
	}
	*ca = SelfSigned{
		Certificate: certificate,
		PrivateKey:  privateKey,
	}

	return
}

// Sign returns an authorized certificate from an x509 struct
//   - template is a populated x509 struct
//   - publicKey is the key from a key-pair that will be
//     inserted into the new certificate
func (ca *SelfSigned) Sign(
	template *x509.Certificate,
	publicKey crypto.PublicKey,
) (certDER parl.CertificateDer, err error) {

	// get certificate authority x509.Certificate
	var caCert *x509.Certificate
	if caCert, err = ca.Validate(); err != nil {
		return
	}
	var caSigner = ca.PrivateKey

	// sign template
	if certDER, err = x509.CreateCertificate(
		rand.Reader,
		template,
		caCert,
		publicKey,
		caSigner,
	); err != nil {
		err = perrors.Errorf("x509.CreateCertificate: '%w'", err)
	}

	return
}

// Validate validates the certificate authority and returns x509 struct format
func (ca *SelfSigned) Validate() (cert *x509.Certificate, err error) {
	if err = ca.PrivateKey.Validate(); err != nil {
		return
	} else if cert, err = ca.ParseCertificate(); perrors.IsPF(&err, "x509.ParseCertificate: ‘%w’", err) {
		return
	} else if cert.PublicKey == nil {
		err = perrors.NewPF("public key uninitialized")
	}
	return
}

// Private returns the certificate authority’s private key
func (ca *SelfSigned) Private() (privateKey parl.PrivateKey) { return ca.PrivateKey }
