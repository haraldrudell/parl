/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"crypto"
	"crypto/x509"
)

// Certificate is container for binary representation of [x509.Certificate]
//   - an implementation is [github.com/haraldrudell/parl/parlca.Certificate]
//   - [x509.Certificate] is struct with field
//     [x509.Certificate.Raw] containing untyped binary DER bytes
//   - [CertificateAuthority.Sign] creates certificates
//   - [github.com/haraldrudell/parl/parlca.LoadFromPem] reads PEM certificate from a file
type Certificate interface {
	// DER returns a typed byte-slice, the most efficient storage format
	//	- der: Distinguished Encoding Rules binary format
	DER() (der CertificateDer)
	// PEM returns text representation: “-----BEGIN CERTIFICATE …”
	//	- private enhanced mail
	PEM() (pemBytes PemBytes)
	// ParseCertificate returns the x509 struct for the certificate
	ParseCertificate() (certificate *x509.Certificate, err error)
}

// CertificateAuthority is an authority that can sign x509 struct certificates
//   - an implementation is [github.com/haraldrudell/parl/parlca/SelfSigned] with certificate and private key
//   - [github.com/haraldrudell/parl/parlca.NewSelfSigned] creates a self-signed certificate authority
//   - [github.com/haraldrudell/parl/parlca.LoadFromPem] reads PEM certificate authority from a file
type CertificateAuthority interface {
	// Validate validates the certificate authority and returns x509 struct format
	Validate() (cert *x509.Certificate, err error)
	// DER returns a typed byte-slice, the most efficient storage format
	//	- der: Distinguished Encoding Rules binary format
	DER() (certificateDer CertificateDer)
	// Sign returns an authorized certificate from an x509 struct
	Sign(template *x509.Certificate, publicKey crypto.PublicKey) (certDER CertificateDer, err error)
	// PEM returns text representation: “-----BEGIN CERTIFICATE …”
	//	- private enhanced mail
	PEM() (pemBytes PemBytes)
	// Private returns the certificate authority’s private key
	Private() (privateKey PrivateKey)
}

// PrivateKey implements [crypto.Signer] and can be used as [tls.Certificate.PrivateKey]
//   - [github.com/haraldrudell/parl/parlca.LoadFromPem] reads PEM certificate authority from a file
//   - [PrivateKeyFactory.NewPrivateKey] is abtract key factory
//   - — [github.com/haraldrudell/parl/parlca.NewPrivateKey] creates private keys
//   - implementation depends on algorithm
type PrivateKey interface {
	// Signer provides singing operations, methods: Public() Sign()
	crypto.Signer
	// DER returns typed key material PKCS#8:, both private and public keys
	//	- PKCS#8 may be encrypted
	//	- der: Distinguished Encoding Rules binary format
	DER() (privateKeyDer PrivateKeyDer, err error)
	// DER returns typed key material, both private and public keys panic on error
	//	- der: Distinguished Encoding Rules binary format
	DERe() (privateKeyDer PrivateKeyDer)
	// PEM returns text representation: “-----BEGIN PRIVATE KEY …”
	//	- private enhanced mail
	PEM() (pemBytes PemBytes, err error)
	// PEMe returns text representation: “-----BEGIN PRIVATE KEY …” panic on error
	//	- private enhanced mail
	PEMe() (pemBytes PemBytes)
	// PublicKey returns the public key from the key-pair
	PublicKey() (publicKey PublicKey)
	// Algo returns private key algorithm
	Algo() (algo x509.PublicKeyAlgorithm)
	// Validate ensures the private key is present, modeled after [rsa.Validate]
	Validate() (err error)
}

// PublicKey contains a public key extracted from a KeyPair
//   - implementation depends on algorithm
type PublicKey interface {
	// DER returns typed key material PKCS#8: both private and public keys
	//	- PKCS#8 may be encrypted
	//	- der: Distinguished Encoding Rules binary format
	DER() (publicKeyDer PublicKeyDer, err error)
	// DER returns typed key material PKCS#8: both private and public keys panic on error
	//	- der: Distinguished Encoding Rules binary format
	DERe() (publicKeyDer PublicKeyDer)
	// PEM returns text representation: “-----BEGIN CERTIFICATE …”
	//	- private enhanced mail
	PEM() (pemBytes PemBytes, err error)
	// PEMe returns text representation: “-----BEGIN CERTIFICATE …” panic on error
	//	- private enhanced mail
	PEMe() (pemBytes PemBytes)
	// Equal compares two public keys
	Equal(x crypto.PublicKey) (isEqual bool)
	// Algo returns public key algorithm
	Algo() (algo x509.PublicKeyAlgorithm)
}

// CertificateDer is a binary container for certificate.
//   - der: Distinguished Encoding Rules is a binary format based on asn1.
type CertificateDer []byte

// PublicKeyDer is a binary container for public key
//   - der: Distinguished Encoding Rules is a binary format based on asn1.
type PublicKeyDer []byte

// PrivateKeyDer is a binary container for private and public key-pair
//   - der: Distinguished Encoding Rules is a binary format based on asn1.
type PrivateKeyDer []byte

// PemBytes bytes is 7-bit ascii string representing keys or certificates
//   - “-----BEGIN CERTIFICATE …”
//   - private enhanced mail
type PemBytes []byte

// PrivateKeyFactory is abstract private key factory
type PrivateKeyFactory interface {
	// NewPrivateKey generates private key for algorithm algo
	//	- [x509.RSA] [x509.ECDSA] [x509.Ed25519]
	NewPrivateKey(algo x509.PublicKeyAlgorithm) (privateKey PrivateKey, err error)
}
