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
	"github.com/haraldrudell/parl/parlca/calib"
	"github.com/haraldrudell/parl/perrors"
)

// EcdsaPrivateKey wraps a binary ecdsa key-pair with methods
type EcdsaPrivateKey struct {
	*ecdsa.PrivateKey
}

// EcdsaPrivateKey is [parl.PrivateKey] includes [crypto.Signer]
var _ parl.PrivateKey = &EcdsaPrivateKey{}

// NewEcdsa
//   - —
//   - returns pointer beacuse ecdsa can only generate pointer
func NewEcdsa(fieldp ...*EcdsaPrivateKey) (privateKey *EcdsaPrivateKey, err error) {

	// get privateKey
	if len(fieldp) > 0 {
		privateKey = fieldp[0]
	}
	if privateKey == nil {
		// allocation here
		privateKey = &EcdsaPrivateKey{}
	}

	// P-256 is 128 bit security
	if privateKey.PrivateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); err != nil {
		err = perrors.ErrorfPF("ecdsa.GenerateKey %w", err)
	}

	return
}

// Algo returns algorithm [x509.ECDSA] “ECDSA”
func (key *EcdsaPrivateKey) Algo() (algo x509.PublicKeyAlgorithm) { return x509.ECDSA }

// DER returns PKCS#8 binary form
func (key *EcdsaPrivateKey) DER() (bytes parl.PrivateKeyDer, err error) {
	if bytes, err = x509.MarshalPKCS8PrivateKey(&key.PrivateKey); err != nil {
		err = perrors.Errorf("x509.MarshalPKCS8PrivateKey: '%w'", err)
	}
	return
}

// DERe returns PKCS#8 with panic on error
func (key *EcdsaPrivateKey) DERe() (privateKeyDer parl.PrivateKeyDer) {
	var err error
	if privateKeyDer, err = key.DER(); err != nil {
		panic(err)
	}
	return
}

// PEM returns text format “-----PRIVATE KEY …”
func (key *EcdsaPrivateKey) PEM() (pemBytes parl.PemBytes, err error) {
	block := pem.Block{
		Type: pemPrivateKeyType,
	}
	if block.Bytes, err = key.DER(); err != nil {
		return
	}
	pemBytes = append([]byte(calib.PemText(block.Bytes)), pem.EncodeToMemory(&block)...)
	return
}

// PEMe returns text format “-----PRIVATE KEY …” panic on error
func (key *EcdsaPrivateKey) PEMe() (pemBytes parl.PemBytes) {
	var err error
	if pemBytes, err = key.PEM(); err != nil {
		panic(err)
	}
	return
}

// PublicKey returns the key-pair’s public key
func (key *EcdsaPrivateKey) PublicKey() (publicKey parl.PublicKey) {
	return &EcdsaPublicKey{PublicKey: key.PrivateKey.PublicKey}
}

// Validate ensures a private key is present
func (key *EcdsaPrivateKey) Validate() (err error) {
	if key.PrivateKey.D == nil {
		err = perrors.New("Uninitialized ecdsa private key")
	}

	return
}
