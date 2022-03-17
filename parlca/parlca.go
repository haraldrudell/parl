/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

/*
Package parlca provides a self-signed certificate authority
*/
package parlca

import (
	"crypto"
	"crypto/x509"
	"io"
)

type Certificate interface {
	DER() (der CertificateDER)
}

type CertificateDER []byte
type KeyDER []byte

// PrivateKey does not contain public part of a key pair, only the private key
type PrivateKey interface {
	HasKey() (hasKey bool) // has key material
	Algo() (algo x509.PublicKeyAlgorithm)
	PrivateBytes() (bytes []byte) // untyped private key material
}

// KeyPair implements crypto.Signer and can therefore be used as tls.Certificate.PrivateKey
type KeyPair interface {
	PrivateKey
	Bytes() (bytes KeyDER, err error) // untyped key material, both private and public keys
	PublicBytes() (bytes []byte)      // untyped public key material
	Private() (signer crypto.Signer)  // typed key material implementing crypto.Signer for x509.CreateCertificate and tls.Certificate.PrivateKey
}

type CertificateAuthority interface {
	Check() (isValid bool, cert *x509.Certificate, err error) // gets x509.Certificate version
	DER() (bytes CertificateDER)                              // untyped bytes, der: Distinguished Encoding Rules binary format
	Sign(template *x509.Certificate, publicKey crypto.PublicKey) (certDER CertificateDER, err error)
	SetReader(reader io.Reader)
}

type KeyGenerator func(io.Reader) (keyPair KeyPair, err error) // creates private key, returns public key
