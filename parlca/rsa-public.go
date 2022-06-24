/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/haraldrudell/parl"
)

type RsaPublicKey struct {
	rsa.PublicKey
}

func (key *RsaPublicKey) Algo() (algo x509.PublicKeyAlgorithm) {
	return x509.RSA
}

func (key *RsaPublicKey) DER() (publicKeyDer parl.PublicKeyDer, err error) {
	publicKeyDer = x509.MarshalPKCS1PublicKey(&key.PublicKey)
	return
}

func (key *RsaPublicKey) DERe() (publicKeyDer parl.PublicKeyDer) {
	var err error
	if publicKeyDer, err = key.DER(); err != nil {
		panic(err)
	}
	return
}

func (key *RsaPublicKey) PEM() (pemBytes parl.PemBytes, err error) {
	block := pem.Block{
		Type: pemPublicKeyType,
	}
	if block.Bytes, err = key.DER(); err != nil {
		return
	}
	pemBytes = append([]byte(PemText()), pem.EncodeToMemory(&block)...)
	return
}

func (key *RsaPublicKey) PEMe() (pemBytes parl.PemBytes) {
	var err error
	if pemBytes, err = key.PEM(); err != nil {
		panic(err)
	}
	return
}
