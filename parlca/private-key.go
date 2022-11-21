/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/x509"
	"os"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/punix"
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

func NewPrivateKey2(algo x509.PublicKeyAlgorithm, privateKeyDer parl.PrivateKeyDer) (privateKey parl.PrivateKey, err error) {
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

func LoadPrivateKeyFromDer(filename string, algo x509.PublicKeyAlgorithm, allowNotFound ...bool) (privateKey parl.PrivateKey, err error) {
	allowNotFound0 := len(allowNotFound) > 0 && allowNotFound[0]
	var privateKeyDer parl.PrivateKeyDer
	if privateKeyDer, err = ReadFile(filename, allowNotFound0); err != nil {
		return // file read error return
	} else if allowNotFound0 && privateKeyDer == nil {
		return
	}
	if privateKey, err = NewPrivateKey2(algo, privateKeyDer); err != nil {
		return
	}
	// TODO 220624 validate privateKey?
	return
}

func LoadFromPem(filename string, allowNotFound ...bool) (
	certificate parl.Certificate, privateKey parl.PrivateKey, publicKey parl.PublicKey,
	err error) {
	allowNotFound0 := len(allowNotFound) > 0 && allowNotFound[0]
	var pemBytes parl.PemBytes
	if pemBytes, err = ReadFile(filename, allowNotFound0); err != nil {
		return // file read error return
	} else if allowNotFound0 && pemBytes == nil {
		return
	}
	// TODO 220624 validate privateKey?
	return ParsePEM(pemBytes)
}

func ReadFile(filename string, allowNotFound bool) (byts []byte, err error) {
	if byts, err = os.ReadFile(filename); err != nil {
		if allowNotFound && punix.IsENOENT(err) {
			err = nil
			return // cert file does not exist: byts == nil, err == nil
		}
		perrors.IsPF(&err, "os.ReadFile %w", err)
	}
	return
}
