/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/x509"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pos"
)

// /usr/local/opt/openssl/bin/openssl x509 -in cert.der -inform der -noout -text
// openssl x509 -in /etc/ssl/certs/VeriSign_Universal_Root_Certification_Authority.pem -inform pem -noout -text

const (
	ssDerExt                     = ".der"
	ssPemExt                     = ".pem"
	writeFileModeUrw fs.FileMode = 0600
	openssl                      = "/opt/homebrew/Cellar/openssl@1.1/1.1.1o/bin/openssl"
)

func TestNewSelfSigned(t *testing.T) {
	// doWriteFiles writes keys and certificates to user’s home directory
	doWriteFiles := false
	writeDir := pos.UserHomeDir()

	var err error
	var privateKey parl.PrivateKey
	var x509Certificate *x509.Certificate

	/*
		// what data types must be provided?
		var _ http.Server     // a golang http server is instantiated using http.Server struct
		var _ tls.Config      // tls is defined in the http.Server field TLSConfig *tls.Config, struct
		var _ tls.Certificate // the tls.Config field is Certificates []Certificate, struct
		// tls.Certificate field Certificate [][]byte
		var _ crypto.PrivateKey // tls.Certificate field PrivateKey crypto.PrivateKey: interface{}
		var _ pkix.Name
		var _ x509.Certificate
	*/

	for _, algo := range []x509.PublicKeyAlgorithm{x509.Ed25519, x509.RSA, x509.ECDSA} {

		// create private and public key
		if privateKey, err = NewPrivateKey(algo); err != nil {
			t.Errorf("NewPrivateKey %s %s", algo.String(), perrors.Short(err))
			t.FailNow()
		}

		if doWriteFiles {
			algoName := strings.ToLower(algo.String())

			filename := filepath.Join(writeDir, "ca-"+algoName+"-private"+ssDerExt)
			t.Logf("Writing: %s", filename)
			os.WriteFile(filename, privateKey.DERe(), writeFileModeUrw)
			t.Logf("%s pkey -inform DER -in %s -text -noout", openssl, filename)

			filename = filepath.Join(writeDir, "ca-"+algoName+"-private"+ssPemExt)
			t.Logf("Writing: %s", filename)
			os.WriteFile(filename, privateKey.PEMe(), writeFileModeUrw)
			t.Logf("%s pkey -in %s -text -noout", openssl, filename)

			// public der does not work
			filename = filepath.Join(writeDir, "ca-"+algoName+"-public"+ssDerExt)
			t.Logf("Writing: %s", filename)
			os.WriteFile(filename, privateKey.PublicKey().DERe(), writeFileModeUrw)
			t.Logf("%s pkey -inform DER -in %s -text -noout -pubin", openssl, filename)

			filename = filepath.Join(writeDir, "ca-"+algoName+"-public"+ssPemExt)
			t.Logf("Writing: %s", filename)
			os.WriteFile(filename, privateKey.PublicKey().PEMe(), writeFileModeUrw)
			t.Logf("%s pkey -in %s -text -noout -pubin", openssl, filename)
		}

		// create certificate authority
		var ca parl.CertificateAuthority
		if ca, err = NewSelfSigned("", algo); err != nil {
			t.Errorf("NewSelfSigned %s %s ", algo.String(), perrors.Short(err))
		}

		if doWriteFiles {
			filename := filepath.Join(writeDir, "ca-"+strings.ToLower(algo.String())+ssDerExt)
			t.Logf("Writing: %s", filename)
			os.WriteFile(filename, ca.DER(), writeFileModeUrw)
			t.Logf("%s x509 -in %s -inform der -noout -text", openssl, filename)

			filename = filepath.Join(writeDir, "ca-"+strings.ToLower(algo.String())+ssPemExt)
			t.Logf("Writing: %s", filename)
			os.WriteFile(filename, ca.PEM(), writeFileModeUrw)
			t.Logf("%s x509 -in %s -noout -text", openssl, filename)
		}

		// CertificateAuthority.Check
		if x509Certificate, err = ca.Check(); err != nil {
			t.Errorf("ca.Check: %s", perrors.Short(err))
			t.FailNow()
		}
		_ = x509Certificate

	}

	if doWriteFiles {
		t.Fail()
	}
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
