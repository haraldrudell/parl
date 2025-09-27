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
	"slices"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

const (
	RsaDefaultBits = 2048
)

// RsaPrivateKey wraps an RSA key-pair
//   - a composite binary struct
//   - rsa package can only generate pointer
//   - rsa.PrivateKey is multiple fields with no self-referencing pointers,
//     atomics or locks, ie. it can be copied
//   - RsaPrivateKey will be used as [parl.PrivateKey] interface,
//     meaning a pointer to it will be used
//   - deserializing a key with [x509.ParsePKCS8PrivateKey]
//     produces *rsa.PrivateKey
//   - because the private key is always be a pointer, nothing is saved
//     by copying it to a value-field but the copy incurs cost
//   - a fieldp allows RsaPrivateKey to be stack-allocated or
//     a field from a previous allocation
type RsaPrivateKey struct {
	// Decrypt(rand io.Reader, ciphertext []byte, opts crypto.DecrypterOpts) (plaintext []byte, err error)
	// Equal(x crypto.PrivateKey) bool
	// Precompute()
	// Public() crypto.PublicKey
	// Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error)
	// Size() int
	// Validate() error
	//	- 6 pointers 2 slices and int
	*rsa.PrivateKey
}

// RsaPrivateKey is [parl.PrivateKey] includes [crypto.Signer]
var _ parl.PrivateKey = &RsaPrivateKey{}

// NewRsa returns 2048-bit RSA private key-pair
//   - privateKey: [parl.PrivateKey] in typed, efficient binary format with methods
//   - [RsaPrivateKey] implements [crypto.Signer]
//   - —
//   - [RsaPrivateKey.Sign] signs using the key
//   - [RsaPrivateKey.DERe] gets binary representation
//   - [RsaPrivateKey.PEMe] get text representation
//   - because rsa package can only generate pointer,
//     allocation cannot be avoided thus no fieldp
//   - panic on error
func NewRsa(fieldp ...*RsaPrivateKey) (privateKey *RsaPrivateKey) {
	var err error
	if privateKey, err = NewRsaBits(RsaDefaultBits, fieldp...); err != nil {
		panic(err)
	}
	return
}

// NewRsaBits returns RSA private key-pair of specific bit-size
func NewRsaBits(bits int, fieldp ...*RsaPrivateKey) (privateKey *RsaPrivateKey, err error) {

	// get privateKey
	if len(fieldp) > 0 {
		privateKey = fieldp[0]
	}
	if privateKey == nil {
		// allocation here
		privateKey = &RsaPrivateKey{}
	}

	if privateKey.PrivateKey, err = rsa.GenerateKey(rand.Reader, bits); perrors.IsPF(&err, "rsa.GenerateKey: %w", err) {
		return
	}

	return
}

// Algo returns algorithm [x509.RSA] “RSA”
func (k *RsaPrivateKey) Algo() (algo x509.PublicKeyAlgorithm) { return x509.RSA }

// DER returns PKCS#8 binary form
func (k *RsaPrivateKey) DER() (bytes parl.PrivateKeyDer, err error) {
	if bytes, err = x509.MarshalPKCS8PrivateKey(&k.PrivateKey); err != nil {
		err = perrors.Errorf("x509.MarshalPKCS8PrivateKey: '%w'", err)
	}
	return
}

// DERe returns PKCS#8 with panic on error
func (k *RsaPrivateKey) DERe() (privateKeyDer parl.PrivateKeyDer) {
	var err error
	if privateKeyDer, err = k.DER(); err != nil {
		panic(err)
	}
	return
}

// PEM returns text format “-----PRIVATE KEY …”
func (k *RsaPrivateKey) PEM() (pemBytes parl.PemBytes, err error) {
	block := pem.Block{
		Type: pemPrivateKeyType,
	}
	if block.Bytes, err = k.DER(); err != nil {
		return
	}
	pemBytes = pem.EncodeToMemory(&block)
	pemBytes = slices.Insert(pemBytes, 0, []byte(PemText(block.Bytes))...)

	return
}

// PEMe returns text format “-----PRIVATE KEY …” panic on error
func (k *RsaPrivateKey) PEMe() (pemBytes parl.PemBytes) {
	var err error
	if pemBytes, err = k.PEM(); err != nil {
		panic(err)
	}
	return
}

// PublicKey returns the key-pair’s public key
func (k *RsaPrivateKey) PublicKey() (publicKey parl.PublicKey) {
	return &RsaPublicKey{PublicKey: k.PrivateKey.PublicKey}
}

// Validate ensures a private key is present
func (k *RsaPrivateKey) Validate() (err error) {
	if k.PrivateKey.D == nil {
		return perrors.New("rsa private key uninitialized")
	}
	err = k.PrivateKey.Validate()

	return
}
