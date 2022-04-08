/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pos"
)

// /usr/local/opt/openssl/bin/openssl x509 -in cert.der -inform der -noout -text
// openssl x509 -in /etc/ssl/certs/VeriSign_Universal_Root_Certification_Authority.pem -inform pem -noout -text

func TestNewSelfSigned(t *testing.T) {

	// what data types must be provided?
	var _ http.Server     // a golang http server is instantiated using http.Server struct
	var _ tls.Config      // tls is defined in the http.Server field TLSConfig *tls.Config, struct
	var _ tls.Certificate // the tls.Config field is Certificates []Certificate, struct
	// tls.Certificate field Certificate [][]byte
	var _ crypto.PrivateKey // tls.Certificate field PrivateKey crypto.PrivateKey: interface{}
	var _ pkix.Name
	var _ x509.Certificate

	//cert, err := x509.CreateCertificate()
	//var ca CertificateAuthority
	//ca.SelfSigned()
	ca := NewSelfSigned("")
	isValid, cert, err := ca.Check()
	if err != nil {
		t.Errorf("Check: %+v", err)
		return
	}
	if !isValid {
		t.Errorf("SelfSigned not valid")
		return
	}
	filename := filepath.Join(pos.UserHomeDir(), "cert.der")
	bytes := ca.DER()
	t.Logf("writing: %s bytes: %d", filename, len(bytes))
	if err := writeBytes(filename, bytes); err != nil {
		t.Errorf("writeBytes: %+v", err)
		return
	}
	t.Logf("/usr/local/opt/openssl/bin/openssl x509 -in ~/cert.der -inform der -noout -text")
	_ = cert
}

func writeBytes(filename string, bytes []byte) (err error) {
	var file *os.File
	if file, err = os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600); err != nil {
		err = perrors.Errorf("os.OpenFile %q: '%w'", filename, err)
		return
	}
	defer func() {
		if e := file.Close(); e != nil {
			err = perrors.AppendError(err, e)
		}
	}()
	_, err = file.Write(bytes)
	return
}
