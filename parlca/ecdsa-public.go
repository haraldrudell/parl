/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

type EcdsaPublicKey struct {
	ecdsa.PublicKey
}

func (key *EcdsaPublicKey) Algo() (algo x509.PublicKeyAlgorithm) {
	return x509.ECDSA
}

func (key *EcdsaPublicKey) DER() (publicKeyDer parl.PublicKeyDer, err error) {
	publicKeyDer, err = x509.MarshalPKIXPublicKey(&key.PublicKey)
	perrors.IsPF(&err, "x509.MarshalPKIXPublicKey %w", err)
	return
}

func (key *EcdsaPublicKey) DERe() (publicKeyDer parl.PublicKeyDer) {
	var err error
	if publicKeyDer, err = key.DER(); err != nil {
		panic(err)
	}
	return
}

func (key *EcdsaPublicKey) PEM() (pemBytes parl.PemBytes, err error) {
	block := pem.Block{
		Type: pemPublicKeyType,
	}
	if block.Bytes, err = key.DER(); err != nil {
		return
	}
	pemBytes = append([]byte(PemText()), pem.EncodeToMemory(&block)...)
	return
}

func (key *EcdsaPublicKey) PEMe() (pemBytes parl.PemBytes) {
	var err error
	if pemBytes, err = key.PEM(); err != nil {
		panic(err)
	}
	return
}
