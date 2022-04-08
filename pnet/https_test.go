/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/haraldrudell/parl/parlca"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pos"
)

func TestNewHttps(t *testing.T) {
	TCPaddress := "127.0.0.1:0"
	network := ""
	var certDER parlca.CertificateDER
	var signer crypto.Signer
	hp := NewHttps(TCPaddress, network, certDER, signer)
	_ = hp
}

func TestHttpsListen(t *testing.T) {
	TCPaddress := "127.0.0.1:0"
	network := ""
	protocol := "https://"
	URIPattern := "/"   // "/" matches everything
	canonicalName := "" // ca common name, will default to {host}ca
	// /usr/local/opt/openssl/bin/openssl x509 -in ca.der -inform der -noout -text
	caCertFilename := filepath.Join(pos.UserHomeDir(), "ca.der")
	ownerRW := fs.FileMode(0o600)
	// /usr/local/opt/openssl/bin/openssl pkey -in key.der -inform der -text -noout
	keyFilename := filepath.Join(pos.UserHomeDir(), "key.der")
	// /usr/local/opt/openssl/bin/openssl x509 -in cert.der -inform der -noout -text
	certFilename := filepath.Join(pos.UserHomeDir(), "cert.der")
	//commonName := "" // server certificate, will default to {host}
	//subject := "subject"
	IPv4loopback := net.IPv4(127, 0, 0, 1)
	secs := 0 // 10
	sleepDuration := time.Duration(secs) * time.Second

	t.Log("Creating self-signed certificate authority")
	caCert := parlca.NewSelfSigned(canonicalName)
	writeFile(caCertFilename, caCert.DER(), ownerRW, t.Logf)
	var isValid bool
	var caX509 *x509.Certificate
	isValid, caX509, err := caCert.Check()
	if err != nil {
		t.Error(err)
		return
	}
	if !isValid {
		t.Error(perrors.New("ca Check failed"))
		return
	}

	t.Log("Creating server private key")
	serverKey, err := parlca.NewEd25519()
	if err != nil {
		t.Error(perrors.Errorf("server parlca.NewEd25519: '%w'", err))
		return
	}
	serverSigner := serverKey.Private()   // private key for running the server
	serverPublic := serverSigner.Public() // public key for creating server certificate
	var keyDER parlca.KeyDER
	if keyDER, err = serverKey.Bytes(); err != nil {
		t.Error(err)
		return
	}
	writeFile(keyFilename, keyDER, ownerRW, t.Logf)

	t.Log("Creating server certificate")
	template := x509.Certificate{
		IPAddresses: []net.IP{IPv4loopback, net.IPv6loopback},
	}
	parlca.EnsureServer(&template)
	var certDER parlca.CertificateDER
	if certDER, err = caCert.Sign(&template, serverPublic); err != nil {
		t.Errorf("signing server certificate: %+v", err)
		return
	}
	writeFile(certFilename, certDER, ownerRW, t.Logf)

	t.Log("Starting server")
	hp := NewHttps(TCPaddress, network, certDER, serverSigner)
	//hp.HandleFunc(URIPattern, func(w http.ResponseWriter, req *http.Request) {
	hp.HandleFunc(URIPattern, func(w http.ResponseWriter, req *http.Request) {
		t.Logf("server received request from: %s", req.RemoteAddr)
		w.WriteHeader(http.StatusNoContent)
	})
	// /usr/local/opt/openssl/bin/openssl s_client -connect 127.0.0.1:57984
	errCh := hp.Listen()

	// listen for errors
	var errChWg sync.WaitGroup
	errChWg.Add(1)
	go func() {
		defer errChWg.Done()
		t.Log("Reading errCh")
		err, ok := <-errCh
		if !ok {
			return // errCh closed
		}
		t.Errorf("errCh: %+v", err)
		panic(err)
	}()

	var expectServerError bool
	for once := true; once; once = false {

		t.Logf("waiting for server ready")
		isUp, addr := hp.WaitForUp()
		if !isUp {
			t.Log("hp.WaitForUp: Server failed to start")
			expectServerError = true
			break
		}
		t.Logf("Server listening at address: %s", addr.String())
		if sleepDuration != 0 {
			t.Logf("Sleep %d seconds", sleepDuration/time.Second)
			time.Sleep(sleepDuration)
		}

		requestURL := protocol + addr.String()
		t.Logf("Sending https request to %s", requestURL)
		var resp *http.Response
		if resp, err = Get(requestURL, NewTLSConfig(caX509), nil); err != nil {
			expectServerError = true
			t.Errorf("%+v", perrors.Errorf("http.Get URL: %s: '%w'", requestURL, err))
			break // expect there to be a server error
		}

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("http.Get: bad status code from %s: %d expected: %d", requestURL, resp.StatusCode, http.StatusNoContent)
			return
		}
		if err = resp.Body.Close(); err != nil {
			t.Errorf("Body.Close: %+v", err)
			return
		}
		t.Logf("response status code: %d", resp.StatusCode)
	}

	t.Logf("Shutting down server")
	hp.Shutdown()

	errChWg.Wait() // wait for errCh to close

	if expectServerError {
		panic(perrors.New("Server provided no error")) // errCh should have paniced
	}

	t.Logf("Completed successfully")
}

func writeFile(filename string, bytes []byte, mode fs.FileMode, logf func(format string, args ...interface{})) {
	if filename == "" {
		return
	}
	logf("Writing: %s", filename)
	if err := os.WriteFile(filename, bytes, mode); err != nil {
		panic(perrors.Errorf("os.WriteFile: '%w'", err))
	}
}

func NewHttpsTypes() {
	// http TLS methods
	_ = http.ListenAndServeTLS
	var s http.Server
	_ = s.ListenAndServeTLS
	_ = s.ServeTLS // has http/one-off 2 code
	_ = s.Serve

	// generic listener
	var netListener net.Listener // interface, represents a listening file handle
	_ = netListener.Accept       // func (net.Listener).Accept() (net.Conn, error)
	_ = netListener.Addr         // func (net.Listener).Addr() net.Addr
	_ = netListener.Close        // func (net.Listener).Close() error

	// tcp listener
	var tcp net.TCPListener // net.Listen listener type: *net.TCPListener
	_ = tcp                 // has func (*net.TCPListener).Close() error

	// TLS listener
	// tls.NewListener wraps a listener to be TLS
	_ = tls.NewListener // func tls.NewListener(inner net.Listener, config *tls.Config) net.Listener
	// tls.NewListener requires tls.Config
	// srv field TLSConfig *tls.Config, struct

	// tls.Config
	var tlsConfig tls.Config // struct
	//_ = tlsConfig.BuildNameToCertificate // func (*tls.Config).BuildNameToCertificate()
	_ = tlsConfig.Clone                // func (*tls.Config).Clone() *tls.Config
	_ = tlsConfig.SetSessionTicketKeys // func (*tls.Config).SetSessionTicketKeys(keys [][32]byte)
	_ = tlsConfig.Certificates         // field Certificates []tls.Certificate

	// tls.Certificate
	var tlsCertificate tls.Certificate // struct no methods
	_ = tlsCertificate.PrivateKey      // field PrivateKey crypto.PrivateKey
	// tlsCertificate.Certificate is a list of certificate DER bytes
	_ = tlsCertificate.Certificate // field Certificate [][]byte
	_ = tlsCertificate

	// http.Server.ServeTLS loads credentials using tls.LoadX509KeyPair
	_ = tls.LoadX509KeyPair // func tls.LoadX509KeyPair(certFile string, keyFile string) (tls.Certificate, error)
	// tls.LoadX509KeyPair uses tls.X509KeyPair to create tls.Certificate
	_ = tls.X509KeyPair // func tls.X509KeyPair(certPEMBlock []byte, keyPEMBlock []byte) (tls.Certificate, error)
	// tls.X509KeyPair uses x509.ParseCertificate to create certificate from DER
	_ = x509.ParseCertificate              // func x509.ParseCertificate(asn1Data []byte) (*x509.Certificate, error)
	var x509Certificate x509.Certificate   // struct
	_ = x509Certificate.CheckCRLSignature  // func (*x509.Certificate).CheckCRLSignature(crl *pkix.CertificateList) error
	_ = x509Certificate.CheckSignature     // func (*x509.Certificate).CheckSignature(algo x509.SignatureAlgorithm, signed []byte, signature []byte) error
	_ = x509Certificate.CheckSignatureFrom // func (*x509.Certificate).CheckSignatureFrom(parent *x509.Certificate) error
	_ = x509Certificate.CreateCRL          // func (*x509.Certificate).CreateCRL(rand io.Reader, priv interface{}, revokedCerts []pkix.RevokedCertificate, now time.Time, expiry time.Time) (crlBytes []byte, err error)
	_ = x509Certificate.Equal              // func (*x509.Certificate).Equal(other *x509.Certificate) bool
	_ = x509Certificate.Verify             // func (*x509.Certificate).Verify(opts x509.VerifyOptions) (chains [][]*x509.Certificate, err error)
	_ = x509Certificate.VerifyHostname     // func (*x509.Certificate).VerifyHostname(h string) error
	_ = x509Certificate.Raw                // field Raw []byte
	_ = x509Certificate
	// tls.X509KeyPair assigns private key DER to cert.PrivateKey

	// 220220 var tlsDial tls.Dial

	var _ crypto.PrivateKey
}
