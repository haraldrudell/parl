/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

const (
	rsaDefaultBits = 2048
)

type RsaPrivateKey struct {
	// Decrypt(rand io.Reader, ciphertext []byte, opts crypto.DecrypterOpts) (plaintext []byte, err error)
	// Equal(x crypto.PrivateKey) bool
	// Precompute()
	// Public() crypto.PublicKey
	// Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error)
	// Size() int
	// Validate() error
	rsa.PrivateKey
}

func NewRsa() (privateKey parl.PrivateKey, err error) {
	return NewRsaBits(rsaDefaultBits)
}

func NewRsaBits(bits int) (privateKey parl.PrivateKey, err error) {
	var rsaPrivateKey *rsa.PrivateKey
	if rsaPrivateKey, err = rsa.GenerateKey(rand.Reader, bits); perrors.IsPF(&err, "rsa.GenerateKey: %w", err) {
		return
	}
	privateKey = &RsaPrivateKey{PrivateKey: *rsaPrivateKey}
	return
}

func (key *RsaPrivateKey) Algo() (algo x509.PublicKeyAlgorithm) {
	return x509.RSA
}

func (key *RsaPrivateKey) DER() (bytes parl.PrivateKeyDer, err error) {
	if bytes, err = x509.MarshalPKCS8PrivateKey(&key.PrivateKey); err != nil {
		err = perrors.Errorf("x509.MarshalPKCS8PrivateKey: '%w'", err)
	}
	return
}

func (key *RsaPrivateKey) DERe() (privateKeyDer parl.PrivateKeyDer) {
	var err error
	if privateKeyDer, err = key.DER(); err != nil {
		panic(err)
	}
	return
}

func (key *RsaPrivateKey) PEM() (pemBytes parl.PemBytes, err error) {
	block := pem.Block{
		Type: pemPrivateKeyType,
	}
	if block.Bytes, err = key.DER(); err != nil {
		return
	}
	pemBytes = append([]byte(PemText(block.Bytes)), pem.EncodeToMemory(&block)...)
	return
}

func (key *RsaPrivateKey) PEMe() (pemBytes parl.PemBytes) {
	var err error
	if pemBytes, err = key.PEM(); err != nil {
		panic(err)
	}
	return
}

func (key *RsaPrivateKey) PublicKey() (publicKey parl.PublicKey) {
	return &RsaPublicKey{PublicKey: key.PrivateKey.PublicKey}
}

func (key *RsaPrivateKey) Validate() (err error) {
	if key.PrivateKey.D == nil {
		return perrors.New("rsa priovate key uninitialized")
	}
	return key.PrivateKey.Validate()
}
