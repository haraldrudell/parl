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

func ParsePEM(pemData []byte) (certificate parl.Certificate, privateKey parl.PrivateKey, publicKey parl.PublicKey, err error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		err = perrors.NewPF("PEM block not found in input")
		return
	}
	switch block.Type {
	case pemPublicKeyType:
		publicKey, err = ParsePkix(block.Bytes)
		return
	case pemPrivateKeyType:
		privateKey, err = ParsePkcs8(block.Bytes)
		return
	case pemCertificateType:
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
