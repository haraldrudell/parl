/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"io"

	"github.com/haraldrudell/parl/error116"
)

type Ed25519KeyPair struct {
	// func (ed25519.PrivateKey).Equal(x crypto.PrivateKey) bool
	// func (ed25519.PrivateKey).Public() crypto.PublicKey
	// func (ed25519.PrivateKey).Seed() []byte
	// func (ed25519.PrivateKey).Sign(rand io.Reader, message []byte, opts crypto.SignerOpts) (signature []byte, err error)
	// implements crypto.Signer: Public, Sign
	ed25519.PrivateKey // type: []byte
}

func NewEd25519() (keyPair KeyPair, err error) {
	return GenerateEd25519(rand.Reader)
}

func GenerateEd25519(reader io.Reader) (keyPair KeyPair, err error) {
	k := Ed25519KeyPair{}
	if _, k.PrivateKey, err = ed25519.GenerateKey(reader); err != nil {
		err = error116.Errorf("ed25519.GenerateKey: '%w'", err)
		return
	}
	keyPair = &k
	return
}

func (key *Ed25519KeyPair) Algo() (algo x509.PublicKeyAlgorithm) {
	return x509.Ed25519
}

func (key *Ed25519KeyPair) Bytes() (bytes KeyDER, err error) {
	if bytes, err = x509.MarshalPKCS8PrivateKey(key.PrivateKey); err != nil {
		err = error116.Errorf("x509.MarshalPKCS8PrivateKey: '%w'", err)
	}
	return
}

func (key *Ed25519KeyPair) HasKey() (hasKey bool) {
	return len(key.PrivateKey) > 0
}

func (key *Ed25519KeyPair) PublicBytes() (bytes []byte) {
	public := key.Public() // type: crypto.PublicKey: interface{} value: []byte

	// interface requires type assertion
	// ed25519.PublicKey: []byte
	var ok bool
	if bytes, ok = public.(ed25519.PublicKey); !ok {
		panic(error116.Errorf("ed25519.PrivateKey.Public not []byte: %T %#[1]v", public))
	}
	return
}

func (key *Ed25519KeyPair) PrivateBytes() (bytes []byte) {
	return key.Seed()
}

func (key *Ed25519KeyPair) Private() (signer crypto.Signer) {
	return key.PrivateKey
}

/*
TODO write to file
func (key *ed25519PrivateKey) WriteFile(path string) (err error) {
	os.WriteFile(path,key. , os.File)
}
*/
