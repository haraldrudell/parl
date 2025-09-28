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

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// parsePkix parses a public key in PKIX, ASN.1 binary DER form
func parsePkix(publicKeyDer parl.PublicKeyDer) (publicKey parl.PublicKey, err error) {
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
