/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

type EcdsaPrivateKey struct {
	ecdsa.PrivateKey
}

func NewEcdsa() (privateKey parl.PrivateKey, err error) {
	var ecdsaPrivateKey *ecdsa.PrivateKey
	// P-256 is 128 bit security
	if ecdsaPrivateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); perrors.IsPF(&err, "ecdsa.GenerateKey %w", err) {
		return
	}
	privateKey = &EcdsaPrivateKey{PrivateKey: *ecdsaPrivateKey}
	return
}

func (key *EcdsaPrivateKey) Algo() (algo x509.PublicKeyAlgorithm) {
	return x509.ECDSA
}

func (key *EcdsaPrivateKey) DER() (bytes parl.PrivateKeyDer, err error) {
	if bytes, err = x509.MarshalPKCS8PrivateKey(&key.PrivateKey); err != nil {
		err = perrors.Errorf("x509.MarshalPKCS8PrivateKey: '%w'", err)
	}
	return
}

func (key *EcdsaPrivateKey) DERe() (privateKeyDer parl.PrivateKeyDer) {
	var err error
	if privateKeyDer, err = key.DER(); err != nil {
		panic(err)
	}
	return
}

func (key *EcdsaPrivateKey) PEM() (pemBytes parl.PemBytes, err error) {
	block := pem.Block{
		Type: pemPrivateKeyType,
	}
	if block.Bytes, err = key.DER(); err != nil {
		return
	}
	pemBytes = append([]byte(PemText(block.Bytes)), pem.EncodeToMemory(&block)...)
	return
}

func (key *EcdsaPrivateKey) PEMe() (pemBytes parl.PemBytes) {
	var err error
	if pemBytes, err = key.PEM(); err != nil {
		panic(err)
	}
	return
}

func (key *EcdsaPrivateKey) PublicKey() (publicKey parl.PublicKey) {
	return &EcdsaPublicKey{PublicKey: key.PrivateKey.PublicKey}
}

func (key *EcdsaPrivateKey) Validate() (err error) {
	if key.PrivateKey.D == nil {
		err = perrors.New("Uninitialized ecdsa private key")
		return
	}
	return
}
