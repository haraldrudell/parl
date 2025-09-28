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

// parsePkcs8 parses an unencrypted private key in PKCS #8, ASN.1 binary DER form
//   - —
//   - at least two allocations
func parsePkcs8(privateKeyDer parl.PrivateKeyDer) (privateKey parl.PrivateKey, err error) {

	// keyAny is return value from ParsePKCS8PrivateKey
	//	- for rsa end ecdsa an allocated pointer
	var keyAny any
	if keyAny, err = x509.ParsePKCS8PrivateKey(privateKeyDer); perrors.IsPF(&err, "x509.ParsePKCS8PrivateKey %w", err) {
		return
	}
	if pk, ok := keyAny.(*rsa.PrivateKey); ok {
		privateKey = &RsaPrivateKey{PrivateKey: pk}
	} else if pk, ok := keyAny.(*ecdsa.PrivateKey); ok {
		privateKey = &EcdsaPrivateKey{PrivateKey: pk}
	} else if pk, ok := keyAny.(ed25519.PrivateKey); ok {
		privateKey = &Ed25519PrivateKey{PrivateKey: pk}
	} else {
		err = perrors.ErrorfPF("Unknown private key type: %T", keyAny)
	}

	return
}
