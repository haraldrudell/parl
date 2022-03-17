/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/rand"
	"crypto/rsa"
)

func f() {
	var _ rsa.PrivateKey // struct
	// func (*rsa.PrivateKey).Equal(x crypto.PrivateKey) bool
	// func (*rsa.PrivateKey).Public() crypto.PublicKey
	// func (*rsa.PublicKey).Size() int
	// func (*rsa.PrivateKey).Validate() error
	// func (*rsa.PrivateKey).Decrypt(rand io.Reader, ciphertext []byte, opts crypto.DecrypterOpts) (plaintext []byte, err error)
	// func (*rsa.PrivateKey).Precompute()
	// func (*rsa.PrivateKey).Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error)
	// implements crypto.Signer
	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	_ = rsaPrivateKey.Equal
	_ = rsaPrivateKey.Public
	_ = rsaPrivateKey.Size
	_ = rsaPrivateKey.Validate
	_ = rsaPrivateKey.Decrypt
	_ = rsaPrivateKey.Precompute
	_ = rsaPrivateKey.Sign
	_ = err
	panic("rsa NIMP")
}
