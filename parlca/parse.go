/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// ParsePEM reads the content in a pem-format byte sequence
//   - pemData: text pem-format data byte-sequence
//   - certificate: non-nil if the first pem-block successfully parsed as a “CERTIFICATE”
//   - privateKey: non-nil if the first pem contained a pkcs8 “PRIVATE KEY”
//   - publicKey: non-nil if the first pem contained a pkix encoded “PUBLIC KEY”
//   - can do rsa, ecdsa, ed25519 keys and x.509 certificates
//   - reads the first pem-block present
//   - errors:
//   - — no pem-block found
//   - — pem parsing failed
//   - — a different pem block type was encountered
func ParsePEM(pemData []byte) (certificate parl.Certificate, privateKey parl.PrivateKey, publicKey parl.PublicKey, err error) {

	// decode the pem block to obtain its type
	var block, _ = pem.Decode(pemData)
	if block == nil {
		err = perrors.NewPF("PEM block not found in input")
		return
	}
	switch block.Type {
	case pemPublicKeyType:

		// “PUBLIC KEY”
		publicKey, err = ParsePkix(block.Bytes)
		return
	case pemPrivateKeyType:

		// “PRIVATE KEY”
		privateKey, err = ParsePkcs8(block.Bytes)
		return
	case pemCertificateType:

		// “CERTIFICATE”
		if _, err = x509.ParseCertificate(block.Bytes); perrors.IsPF(&err, "x509.ParseCertificate %w", err) {
			return
		}
		certificate = &Certificate{der: block.Bytes}
		return
	default:
		err = perrors.ErrorfPF("Unknown pem block type: %q", block.Type)
		return
	}
}

// ParsePkcs8 parses an unencrypted private key in PKCS #8, ASN.1 binary DER form
func ParsePkcs8(privateKeyDer parl.PrivateKeyDer) (privateKey parl.PrivateKey, err error) {
	var pub any
	if pub, err = x509.ParsePKCS8PrivateKey(privateKeyDer); perrors.IsPF(&err, "x509.ParsePKCS8PrivateKey %w", err) {
		return
	}
	if pk, ok := pub.(*rsa.PrivateKey); ok {
		privateKey = &RsaPrivateKey{PrivateKey: *pk}
	} else if pk, ok := pub.(*ecdsa.PrivateKey); ok {
		privateKey = &EcdsaPrivateKey{PrivateKey: *pk}
	} else if pk, ok := pub.(ed25519.PrivateKey); ok {
		privateKey = &Ed25519PrivateKey{PrivateKey: pk}
	} else {
		err = perrors.ErrorfPF("Unknown private key type: %T", pub)
	}
	return
}

// ParsePkix parses a public key in PKIX, ASN.1 binary DER form
func ParsePkix(publicKeyDer parl.PublicKeyDer) (publicKey parl.PublicKey, err error) {
	var pub any
	if pub, err = x509.ParsePKIXPublicKey(publicKeyDer); perrors.IsPF(&err, "x509.ParsePKIXPublicKey %w", err) {
		return
	}
	if pk, ok := pub.(*rsa.PublicKey); ok {
		publicKey = &RsaPublicKey{PublicKey: *pk}
	} else if pk, ok := pub.(*ecdsa.PublicKey); ok {
		publicKey = &EcdsaPublicKey{PublicKey: *pk}
	} else if pk, ok := pub.(ed25519.PublicKey); ok {
		publicKey = &Ed25519PublicKey{PublicKey: pk}
	} else {
		err = perrors.ErrorfPF("Unknown public key type: %T", pub)
	}
	return
}
