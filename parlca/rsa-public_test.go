/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/x509"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

func TestRsaPublicKey_Algo(t *testing.T) {
	var privateKey parl.PrivateKey
	var err error
	if privateKey, err = NewRsa(); err != nil {
		t.Errorf("NewRsa %s", perrors.Short(err))
		t.FailNow()
	}
	publicKey := privateKey.PublicKey()
	if publicKey.Algo() != x509.RSA {
		t.Errorf("bad algo %s exp %s", publicKey.Algo(), x509.RSA)
	}
}
