/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// Ed25519 implements parl.KeyPair for the x509.Ed25519 algorithm.
type Ed25519PrivateKey struct {
	// func (ed25519.PrivateKey).Equal(x crypto.PrivateKey) bool
	// func (ed25519.PrivateKey).Public() crypto.PublicKey
	// func (ed25519.PrivateKey).Seed() []byte
	// func (ed25519.PrivateKey).Sign(rand io.Reader, message []byte, opts crypto.SignerOpts) (signature []byte, err error)
	// implements crypto.Signer: Public, Sign
	ed25519.PrivateKey // type: []byte
}

var _ crypto.Signer = &ed25519.PrivateKey{}

func NewEd25519() (privateKey parl.PrivateKey, err error) {
	k := Ed25519PrivateKey{}
	if _, k.PrivateKey, err = ed25519.GenerateKey(nil); perrors.IsPF(&err, "ed25519.GenerateKey %w", err) {
		return
	}
	privateKey = &k
	return
}

func (key *Ed25519PrivateKey) Algo() (algo x509.PublicKeyAlgorithm) {
	return x509.Ed25519
}

func (key *Ed25519PrivateKey) DER() (privateKeyDer parl.PrivateKeyDer, err error) {
	if privateKeyDer, err = x509.MarshalPKCS8PrivateKey(key.PrivateKey); err != nil {
		err = perrors.Errorf("x509.MarshalPKCS8PrivateKey: '%w'", err)
	}
	return
}

func (key *Ed25519PrivateKey) DERe() (privateKeyDer parl.PrivateKeyDer) {
	var err error
	if privateKeyDer, err = key.DER(); err != nil {
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

func (key *Ed25519PrivateKey) PEM() (pemBytes parl.PemBytes, err error) {
	block := pem.Block{
		Type: pemPrivateKeyType,
	}
	if block.Bytes, err = key.DER(); err != nil {
		return
	}
	pemBytes = append([]byte(PemText(key.PrivateKey.Seed())), pem.EncodeToMemory(&block)...)
	return
}

func (key *Ed25519PrivateKey) PEMe() (pemBytes parl.PemBytes) {
	var err error
	if pemBytes, err = key.PEM(); err != nil {
		panic(err)
	}
	return
}

func (key *Ed25519PrivateKey) PublicKey() (publicKey parl.PublicKey) {
	return &Ed25519PublicKey{PublicKey: key.PrivateKey.Public().(ed25519.PublicKey)}
}

func (key *Ed25519PrivateKey) Validate() (err error) {
	length := len(key.PrivateKey)
	if length == 0 {
		return perrors.New("ed25519 private key uninitialized")
	}
	if length != ed25519.PrivateKeySize {
		return perrors.New("ed25519 private key corrupt")
	}
	return
}
