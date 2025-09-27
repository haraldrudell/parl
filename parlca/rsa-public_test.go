/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/x509"
	"testing"

	"github.com/haraldrudell/parl"
)

func TestRsaPublicKey_Algo(t *testing.T) {
	var privateKey parl.PrivateKey = NewRsa()
	var publicKey = privateKey.PublicKey()
	if publicKey.Algo() != x509.RSA {
		t.Errorf("bad algo %s exp %s", publicKey.Algo(), x509.RSA)
	}
}
