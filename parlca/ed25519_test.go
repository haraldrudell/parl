/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto"
	"crypto/ed25519"
	"crypto/tls"
	"crypto/x509"
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

func TestNewEd25519(t *testing.T) {
	var ed25519PrivateKey ed25519.PrivateKey // type: []byte
	// implements crypto.Signer: Public, Sign
	_ = ed25519PrivateKey.Equal  // func (ed25519.PrivateKey).Equal(x crypto.PrivateKey) bool
	_ = ed25519PrivateKey.Public // func (ed25519.PrivateKey).Public() crypto.PublicKey
	_ = ed25519PrivateKey.Seed   // func (ed25519.PrivateKey).Seed() []byte
	_ = ed25519PrivateKey.Sign   // func (ed25519.PrivateKey).Sign(rand io.Reader, message []byte, opts crypto.SignerOpts) (signature []byte, err error)
	_ = ed25519PrivateKey

	var cryptoPrivateKey crypto.PrivateKey // interface{} no methods
	_ = cryptoPrivateKey

	var cryptoPublicKey crypto.PublicKey // interface{} no methods
	_ = cryptoPublicKey

	var cryptoSigner ed25519.PrivateKey // interface, ed25519.PrivateKey implements crypto.Signer
	_ = cryptoSigner.Public             // func (crypto.Signer).Public() crypto.PublicKey
	_ = cryptoSigner.Sign               // func (crypto.Signer).Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error)
	_ = cryptoSigner

	var tlsCertiticate tls.Certificate
	_ = tlsCertiticate.PrivateKey // crypto.PrivateKey
	// comment on field PrivateKey crypto.PrivateKey: This must implement crypto.Signer with an RSA, ECDSA or Ed25519 PublicKey

	var _ x509.PublicKeyAlgorithm // int

	var ed25519PublicKey ed25519.PublicKey // []byte
	_ = ed25519PublicKey.Equal             // func (ed25519.PublicKey).Equal(x crypto.PublicKey) bool
	_ = ed25519PublicKey

	keyPair, err := NewEd25519()
	if err != nil {
		t.Error(err)
		return
	}

	if keyPair == nil {
		t.Errorf("NewEd25519 returned nil")
		return
	}
	if !keyPair.HasKey() {
		t.Error(perrors.New("keyPair empty"))
		return
	}
	algo := keyPair.Algo()
	if algo != x509.Ed25519 {
		t.Errorf("Unknown algo: %s", algo.String())
		return
	}
	var keyDER KeyDER
	if keyDER, err = keyPair.Bytes(); err != nil {
		t.Error(err)
		return
	}
	if len(keyDER) == 0 {
		t.Errorf("private key empty")
		return
	}
	bytes := keyPair.PublicBytes()
	if len(bytes) == 0 {
		t.Errorf("public key empty")
		return
	}
}
