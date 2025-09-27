/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/x509"

	"github.com/haraldrudell/parl"
)

// NewPrivateKey returns key-pair for selected algorithm
//   - parlca.NewPrivateKey implements [parl.PrivateKeyFactory] abstract factory
func NewPrivateKey(algo x509.PublicKeyAlgorithm) (privateKey parl.PrivateKey, err error) {
	switch algo {
	case x509.Ed25519:
		var eKey Ed25519PrivateKey
		eKey, err = MakeEd25519()
		privateKey = &eKey
	case x509.RSA:
		privateKey = NewRsa()
	case x509.ECDSA:
		privateKey, err = NewEcdsa()
	default:
		err = x509.ErrUnsupportedAlgorithm
	}
	return
}
