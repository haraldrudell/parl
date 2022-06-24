/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

type Ed25519PublicKey struct {
	ed25519.PublicKey // Equal()
}

var _ parl.PublicKey = &Ed25519PublicKey{}

func (key *Ed25519PublicKey) DER() (publicKeyDer parl.PublicKeyDer, err error) {
	if len(key.PublicKey) == 0 {
		err = perrors.New("Uninitialized ed25519 public key")
		return
	} else if len(key.PublicKey) != ed25519.PublicKeySize {
		err = perrors.Errorf("Bad ed25519 public key length: %d exp: %d", len(key.PublicKey), ed25519.PublicKeySize)
		return
	}
	var byts []byte
	if byts, err = x509.MarshalPKIXPublicKey(key.PublicKey); perrors.IsPF(&err, "x509.MarshalPKIXPublicKey %w", err) {
		return
	}
	publicKeyDer = byts
	return
}

func (key *Ed25519PublicKey) DERe() (publicKeyDer parl.PublicKeyDer) {
	var err error
	if publicKeyDer, err = key.DER(); err != nil {
		panic(err)
	}
	return
}

func (key *Ed25519PublicKey) PEM() (pemBytes parl.PemBytes, err error) {
	block := pem.Block{
		Type: pemPublicKeyType,
	}
	if block.Bytes, err = key.DER(); err != nil {
		return
	}
	pemBytes = append([]byte(PemText()), pem.EncodeToMemory(&block)...)
	return
}

func (key *Ed25519PublicKey) PEMe() (pemBytes parl.PemBytes) {
	var err error
	if pemBytes, err = key.PEM(); err != nil {
		panic(err)
	}
	return
}

func (key *Ed25519PublicKey) Algo() (algo x509.PublicKeyAlgorithm) {
	return x509.Ed25519
}
