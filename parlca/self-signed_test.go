/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca_test

import (
	"crypto/x509"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/parlca"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pos"
)

// /usr/local/opt/openssl/bin/openssl x509 -in cert.der -inform der -noout -text
// openssl x509 -in /etc/ssl/certs/VeriSign_Universal_Root_Certification_Authority.pem -inform pem -noout -text

func TestSelfSigned(t *testing.T) {
	// doWriteFiles writes keys and certificates to user’s home directory
	const doWriteFiles = false

	if doWriteFiles {
		defer t.Errorf("Logging on for write-files")
	}
	const (
		ssDerExt                     = ".der"
		ssPemExt                     = ".pem"
		writeFileModeUrw fs.FileMode = 0600
		openssl                      = "/opt/homebrew/Cellar/openssl@1.1/1.1.1o/bin/openssl"
		canonicalName                = ""
	)
	var (
		writeDir = pos.UserHomeDir()
		algoList = []x509.PublicKeyAlgorithm{x509.Ed25519, x509.RSA, x509.ECDSA}
	)

	var (
		err             error
		privateKey      parl.PrivateKey
		x509Certificate *x509.Certificate
	)

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

	for _, algo := range algoList {

		// create private and public key
		if privateKey, err = parlca.NewPrivateKey(algo); err != nil {
			t.Fatalf("NewPrivateKey %s %s", algo.String(), perrors.Short(err))
		}

		if doWriteFiles {
			var algoName = strings.ToLower(algo.String())
			var derFile = filepath.Join(writeDir, "ca-"+algoName+"-private"+ssDerExt)
			var pemFile = filepath.Join(writeDir, "ca-"+algoName+"-private"+ssPemExt)
			var publicDerFile = filepath.Join(writeDir, "ca-"+algoName+"-public"+ssDerExt)
			var publicPemFile = filepath.Join(writeDir, "ca-"+algoName+"-public"+ssPemExt)

			t.Logf("Writing: %s", derFile)
			pos.WriteNewFile(derFile, privateKey.DERe())
			t.Logf("%s pkey -inform DER -in %s -text -noout", openssl, derFile)

			t.Logf("Writing: %s", pemFile)
			pos.WriteNewFile(pemFile, privateKey.PEMe())
			t.Logf("%s pkey -in %s -text -noout", openssl, pemFile)

			// public der does not work
			t.Logf("Writing: %s", publicDerFile)
			pos.WriteNewFile(publicDerFile, privateKey.PublicKey().DERe())
			t.Logf("%s pkey -inform DER -in %s -text -noout -pubin", openssl, publicDerFile)

			t.Logf("Writing: %s", publicPemFile)
			pos.WriteNewFile(publicPemFile, privateKey.PublicKey().PEMe())
			t.Logf("%s pkey -in %s -text -noout -pubin", openssl, publicPemFile)
		}

		// create certificate authority
		var ca parl.CertificateAuthority
		if ca, err = parlca.NewSelfSigned(canonicalName, algo); err != nil {
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
		if x509Certificate, err = ca.Validate(); err != nil {
			t.Fatalf("ca.Check: %s", perrors.Short(err))
		}
		_ = x509Certificate
	}
}
