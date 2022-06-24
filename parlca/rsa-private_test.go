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

func TestRsa(t *testing.T) {
	var privateKey parl.PrivateKey
	var err error
	if privateKey, err = NewRsa(); err != nil {
		t.Errorf("err NewRsa %s", perrors.Short(err))
		t.FailNow()
	}
	if privateKey.Algo() != x509.RSA {
		t.Errorf("bad algo %s exp %s", privateKey.Algo(), x509.RSA)
	}

	/*
		var rsaPrivateKey *rsa.PrivateKey
		// runtime error: invalid memory address or nil pointer dereference
		if err = rsaPrivateKey.Validate(); err != nil {
			t.Errorf("err Validate %s", perrors.Short(err))
		}
	*/
}
