/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca_test

import (
	"crypto"
	"crypto/x509"
	"testing"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/parlca"
	"github.com/haraldrudell/parl/phttp"
	"github.com/haraldrudell/parl/ptime"
)

func TestCredentials(t *testing.T) {
	//t.Error("Logging on")
	var (
		cert       parl.Certificate
		privateKey crypto.Signer
		err        error
		t0, t1     time.Time
	)

	// Credentials should not return error
	t0 = time.Now()
	cert, privateKey, _ /*caCertDER*/, _ /*caKey*/, err = parlca.CreateCredentials(x509.Ed25519, parlca.DefaultCAName)
	t1 = time.Now()
	if err != nil {
		t.Fatalf("Credentials err “%s”", err)
	}
	// Credentials generated in 2.557ms
	//	- Credentials generated in 1.893ms
	//	- — x509.ParseCertificate costs 2.557 - 1.893 = 0.664 ms
	t.Logf("Credentials generated in %s", ptime.Duration(t1.Sub(t0)))

	// crdentials are tested by [phttp.Https] test
	var _ phttp.Https
	_ = cert
	_ = privateKey
}
