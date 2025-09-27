/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"strings"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

func TestEd25519Public(t *testing.T) {
	const (
		expPem = "-----BEGIN PUBLIC KEY-----\n"
	)

	var (
		publicKey parl.PublicKey
		//var publicKeyDer parl.PublicKeyDer
		pemBytes parl.PemBytes
		keyPair  Ed25519PrivateKey
		err      error
	)

	// get public key
	if keyPair, err = MakeEd25519(); err != nil {
		t.Errorf("Error NewEd25519: %s", perrors.Short(err))
		t.FailNow()
	}
	publicKey = keyPair.PublicKey()

	/*
		// test DER
		publicKeyDer = publicKey.DERe()
		if len(publicKeyDer) != ed25519.PublicKeySize {
			t.Errorf("Bad len public: %d exp %d", len(publicKeyDer), ed25519.PublicKeySize)
		}
	*/
	// test PEM
	pemBytes = publicKey.PEMe()
	pemString := string(pemBytes)

	// pem: "-----BEGIN PUBLIC KEY-----\ni1lp/BZZ8nyCjAe6c1Xj8DzP6Pnc8ApFXbgdrvrXDy0=\n-----END PUBLIC KEY-----\n"
	//t.Logf("pem: %q", s)

	if !strings.Contains(pemString, expPem) {
		t.Errorf("bad pem string: %q", pemString)
	}
}
