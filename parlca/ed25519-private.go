/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"slices"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// Ed25519 is efficient binary container for Ed25519 key-pair with wrapped methods
type Ed25519PrivateKey struct {
	// func (ed25519.PrivateKey).Equal(x crypto.PrivateKey) bool
	// func (ed25519.PrivateKey).Public() crypto.PublicKey
	// func (ed25519.PrivateKey).Seed() []byte
	// func (ed25519.PrivateKey).Sign(rand io.Reader, message []byte, opts crypto.SignerOpts) (signature []byte, err error)
	// implements crypto.Signer: Public, Sign
	//	- type: []byte
	ed25519.PrivateKey
}

// Ed25519PrivateKey is [parl.PrivateKey] includes [crypto.Signer]
var _ parl.PrivateKey = &Ed25519PrivateKey{}

// MakeEd25519 returns Ed25519 key-pair implementing [parl.PrivateKey] and [cypto.Signer]
//   - [Ed25519PrivateKey.Sign] signs using the key
//   - [Ed25519PrivateKey.DERe] gets binary representation
//   - [Ed25519PrivateKey.PEMe] get text representation
func MakeEd25519() (privateKey Ed25519PrivateKey, err error) {
	var k = Ed25519PrivateKey{}
	if _, k.PrivateKey, err = ed25519.GenerateKey(nil); perrors.IsPF(&err, "ed25519.GenerateKey %w", err) {
		return
	}
	privateKey = k
	return
}

// Algo returns algorithm [x509.Ed25519] “Ed25519”
func (e *Ed25519PrivateKey) Algo() (algo x509.PublicKeyAlgorithm) { return x509.Ed25519 }

// DER returns PKCS#8 binary form
func (e *Ed25519PrivateKey) DER() (privateKeyDer parl.PrivateKeyDer, err error) {
	if privateKeyDer, err = x509.MarshalPKCS8PrivateKey(e.PrivateKey); err != nil {
		err = perrors.Errorf("x509.MarshalPKCS8PrivateKey: '%w'", err)
	}
	return
}

// DERe returns PKCS#8 with panic on error
func (e *Ed25519PrivateKey) DERe() (privateKeyDer parl.PrivateKeyDer) {
	var err error
	if privateKeyDer, err = e.DER(); err != nil {
		panic(err)
	}
	return
}

/*
func (key *Ed25519PrivateKey) HasKey() (hasKey bool) {
	return len(key.PrivateKey) > 0
}

func (key *Ed25519PrivateKey) HasPublicKey() (hasPublicKey bool) {
	return len(key.PrivateKey) > 0
}

func (key *Ed25519PrivateKey) Public() (publicKey parl.PublicKey) {
	publicKey = &Ed25519Public{PublicKey: key.PrivateKey.Public().(ed25519.PublicKey)}
	return
}

func (key *Ed25519PrivateKey) PrivateBytes() (bytes []byte) {
	return key.Seed()
}
*/

// PEM returns text format “-----PRIVATE KEY …”
func (e *Ed25519PrivateKey) PEM() (pemBytes parl.PemBytes, err error) {
	block := pem.Block{
		Type: pemPrivateKeyType,
	}
	if block.Bytes, err = e.DER(); err != nil {
		return
	}
	// allocation here
	pemBytes = pem.EncodeToMemory(&block)
	// allocation here
	var seed = e.PrivateKey.Seed()
	var text = PemText(seed)
	pemBytes = slices.Insert(pemBytes, 0, []byte(text)...)
	return
}

// PEMe returns text format “-----PRIVATE KEY …” panic on error
func (e *Ed25519PrivateKey) PEMe() (pemBytes parl.PemBytes) {
	var err error
	if pemBytes, err = e.PEM(); err != nil {
		panic(err)
	}
	return
}

// PublicKey returns the key-pair’s public key
func (e *Ed25519PrivateKey) PublicKey() (publicKey parl.PublicKey) {
	return &Ed25519PublicKey{PublicKey: e.PrivateKey.Public().(ed25519.PublicKey)}
}

// Validate ensures a private key is present
func (e *Ed25519PrivateKey) Validate() (err error) {
	if length := len(e.PrivateKey); length == 0 {
		err = perrors.New("ed25519 private key uninitialized")
	} else if length != ed25519.PrivateKeySize {
		err = perrors.New("ed25519 private key corrupt")
	}
	return
}
