/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"crypto"
	"crypto/x509"
)

type Certificate interface {
	DER() (der CertificateDer)
	PEM() (pemBytes PemBytes)
	ParseCertificate() (certificate *x509.Certificate, err error)
}

type CertificateAuthority interface {
	Check() (cert *x509.Certificate, err error) // gets x509.Certificate version
	DER() (certificateDer CertificateDer)       // untyped bytes, der: Distinguished Encoding Rules binary format
	Sign(template *x509.Certificate, publicKey crypto.PublicKey) (certDER CertificateDer, err error)
	PEM() (pemBytes PemBytes)
}

// PrivateKey implements crypto.Signer and can therefore be used as tls.Certificate.PrivateKey
type PrivateKey interface {
	crypto.Signer                                  // Public() Sign()
	DER() (privateKeyDer PrivateKeyDer, err error) // untyped key material, both private and public keys
	DERe() (privateKeyDer PrivateKeyDer)
	PEM() (pemBytes PemBytes, err error)
	PEMe() (pemBytes PemBytes)
	PublicKey() (publicKey PublicKey)
	Algo() (algo x509.PublicKeyAlgorithm)
	// validate ensures the private key is present, modeled after rsa.Validate
	Validate() (err error)
}

// PublicKey contains a public key extracted from a KeyPair
type PublicKey interface {
	DER() (publicKeyDer PublicKeyDer, err error)
	DERe() (publicKeyDer PublicKeyDer)
	PEM() (pemBytes PemBytes, err error)
	PEMe() (pemBytes PemBytes)
	Equal(x crypto.PublicKey) (isEqual bool)
	Algo() (algo x509.PublicKeyAlgorithm)
}

// CertificateDer is a binary encoding of a certificate.
// der: Distinguished Encoding Rules is a binary format based on asn1.
type CertificateDer []byte

// PublicKeyDer is a binary encoding of a public key
type PublicKeyDer []byte

// PublicKeyDer is a binary encoding of a private and public key-pair
type PrivateKeyDer []byte

// PemBytes bytes is 7-bit ascii string representing keys or certificates
type PemBytes []byte

type PrivateKeyFactory interface {
	NewPrivateKey(algo x509.PublicKeyAlgorithm) (privateKey PrivateKey, err error)
}
