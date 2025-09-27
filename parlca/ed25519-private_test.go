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

	"github.com/haraldrudell/parl"
)

func TestNewEd25519(t *testing.T) {

	var (
		// type: []byte
		ed25519PrivateKey ed25519.PrivateKey
		// implements crypto.Signer: Public, Sign
		// func (ed25519.PrivateKey).Equal(x crypto.PrivateKey) bool
		_ = ed25519PrivateKey.Equal
		// func (ed25519.PrivateKey).Public() crypto.PublicKey
		_ = ed25519PrivateKey.Public
		// func (ed25519.PrivateKey).Seed() []byte
		_ = ed25519PrivateKey.Seed
		// func (ed25519.PrivateKey).Sign(rand io.Reader, message []byte, opts crypto.SignerOpts) (signature []byte, err error)
		_ = ed25519PrivateKey.Sign
		_ = ed25519PrivateKey

		// interface{} no methods
		cryptoPrivateKey crypto.PrivateKey
		_                = cryptoPrivateKey
		// interface{} no methods
		cryptoPublicKey crypto.PublicKey
		_               = cryptoPublicKey
		// interface, ed25519.PrivateKey implements crypto.Signer
		cryptoSigner ed25519.PrivateKey
		// func (crypto.Signer).Public() crypto.PublicKey
		_ = cryptoSigner.Public
		// func (crypto.Signer).Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error)
		_ = cryptoSigner.Sign
		_ = cryptoSigner

		tlsCertiticate tls.Certificate
		// crypto.PrivateKey
		_ = tlsCertiticate.PrivateKey
		// comment on field PrivateKey crypto.PrivateKey: This must implement crypto.Signer with an RSA, ECDSA or Ed25519 PublicKey
		// int
		_ x509.PublicKeyAlgorithm

		// []byte
		ed25519PublicKey ed25519.PublicKey
		// func (ed25519.PublicKey).Equal(x crypto.PublicKey) bool
		_ = ed25519PublicKey.Equal
		_ = ed25519PublicKey
	)

	var (
		keyPair Ed25519PrivateKey
		err     error
		algo    x509.PublicKeyAlgorithm
	)

	keyPair, err = MakeEd25519()
	if err != nil {
		t.Error(err)
		return
	}

	algo = keyPair.Algo()
	if algo != x509.Ed25519 {
		t.Errorf("Unknown algo: %s", algo.String())
		return
	}
	var keyDER parl.PrivateKeyDer
	if keyDER, err = keyPair.DER(); err != nil {
		t.Error(err)
		return
	}
	if len(keyDER) == 0 {
		t.Errorf("private key empty")
		return
	}
	publicKey := keyPair.PublicKey()

	if len(publicKey.DERe()) == 0 {
		t.Errorf("public key empty")
		return
	}

	var key Ed25519PrivateKey
	if err = key.Validate(); err == nil {
		t.Errorf("Mising expected error: Ed25519PrivateKey.Validate")
	}
}
