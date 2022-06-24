/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
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

func NewPrivateKey(algo x509.PublicKeyAlgorithm) (privateKey parl.PrivateKey, err error) {
	switch algo {
	case x509.Ed25519:
		privateKey, err = NewEd25519()
	case x509.RSA:
		privateKey, err = NewRsa()
	case x509.ECDSA:
		privateKey, err = NewEcdsa()
	default:
		err = x509.ErrUnsupportedAlgorithm
	}
	return
}

func parsePEM(pemData []byte) (cert parl.Certificate, private parl.PrivateKey, public parl.PublicKey, err error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		err = perrors.NewPF("PEM block not found in input")
		return
	}
	switch block.Type {
	case pemPublicKeyType:
		var pub any
		if pub, err = x509.ParsePKIXPublicKey(block.Bytes); perrors.IsPF(&err, "x509.ParsePKIXPublicKey %w", err) {
			return
		}
		if pk, ok := pub.(*rsa.PublicKey); ok {
			public = &RsaPublicKey{PublicKey: *pk}
		} else if pk, ok := pub.(*ecdsa.PublicKey); ok {
			public = &EcdsaPublicKey{PublicKey: *pk}
		} else if pk, ok := pub.(ed25519.PublicKey); ok {
			public = &Ed25519PublicKey{PublicKey: pk}
		} else {
			err = perrors.ErrorfPF("Unknown public key type: %T", pub)
		}
		return
	case pemPrivateKeyType:
		var pub any
		if pub, err = x509.ParsePKCS8PrivateKey(block.Bytes); perrors.IsPF(&err, "x509.ParsePKCS8PrivateKey %w", err) {
			return
		}
		if pk, ok := pub.(*rsa.PrivateKey); ok {
			private = &RsaPrivateKey{PrivateKey: *pk}
		} else if pk, ok := pub.(*ecdsa.PrivateKey); ok {
			private = &EcdsaPrivateKey{PrivateKey: *pk}
		} else if pk, ok := pub.(ed25519.PrivateKey); ok {
			private = &Ed25519PrivateKey{PrivateKey: pk}
		} else {
			err = perrors.ErrorfPF("Unknown private key type: %T", pub)
		}
		return
	case pemCertificateType:
		if _, err = x509.ParseCertificate(block.Bytes); perrors.IsPF(&err, "x509.ParseCertificate %w", err) {
			return
		}
		cert = &Certificate{der: block.Bytes}
		return
	default:
		err = perrors.ErrorfPF("Unknown pem block type: %q", block.Type)
		return
	}
}
