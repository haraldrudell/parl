/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phttp

import (
	"crypto"
	"crypto/x509"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/parlca"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pnet"
)

func TestNewHttps(t *testing.T) {
	var ip = "127.0.0.1"
	var nearSocket = netip.MustParseAddrPort(ip + ":0")
	var network = pnet.NetworkTCP
	var addrExp = ip + ":" + strconv.Itoa(int(HttpsPort))

	var certDER parl.CertificateDer
	var signer crypto.Signer
	var https = NewHttps(nearSocket, network, certDER, signer)
	if https.Network != pnet.NetworkTCP {
		t.Errorf("New bad network %s exp %s", https.Network, network)
	}
	if https.Server.Addr != addrExp {
		t.Errorf("bad Addr %q exp %q", https.Server.Addr, addrExp)
	}
}

func TestHttpsListen(t *testing.T) {
	var IPv4loopback = net.IPv4(127, 0, 0, 1)
	var ipS = IPv4loopback.String()
	var dir = t.TempDir()
	var nearSocket = netip.MustParseAddrPort(ipS + ":0")
	var network = pnet.NetworkDefault
	// https://
	var protocol = "https://"
	// "/" matches everything
	var URIPattern = "/"
	// ca common name, will default to {host}ca
	var canonicalName = ""
	// /usr/local/opt/openssl/bin/openssl x509 -in ca.der -inform der -noout -text
	var caCertFilename = filepath.Join(dir, "ca.der")
	var ownerRW = fs.FileMode(0o600)
	// /usr/local/opt/openssl/bin/openssl pkey -in key.der -inform der -text -noout
	var keyFilename = filepath.Join(dir, "key.der")
	// /usr/local/opt/openssl/bin/openssl x509 -in cert.der -inform der -noout -text
	var certFilename = filepath.Join(dir, "cert.der")
	// var commonName = "" // server certificate, will default to {host}
	// var subject = "subject"

	var err error
	var caCert parl.CertificateAuthority
	var caX509 *x509.Certificate
	var serverKey parl.PrivateKey
	// private key for running the server
	var serverSigner parl.PrivateKey
	// public key for creating server certificate
	var serverPublic crypto.PublicKey
	var keyDER parl.PrivateKeyDer
	var template x509.Certificate
	var certDER parl.CertificateDer
	var handler *sHandler
	var goResult = parl.NewGoResult()
	var near, respS string
	var resp *http.Response
	var statusCode int

	t.Log("Creating self-signed certificate authority")
	if caCert, err = parlca.NewSelfSigned(canonicalName, x509.RSA); err != nil {
		t.Fatalf("parlca.NewSelfSigned %s %s", x509.RSA, perrors.Short(err))
	}
	writeFile(caCertFilename, caCert.DER(), ownerRW, t.Logf)
	caX509, err = caCert.Check()
	if err != nil {
		t.Fatalf("caCert,Check: %s", perrors.Short(err))
	}

	t.Log("Creating server private key")
	serverKey, err = parlca.NewEd25519()
	if err != nil {
		t.Fatal(perrors.Errorf("server parlca.NewEd25519: '%w'", err))
	}
	serverSigner = serverKey             // private key for running the server
	serverPublic = serverSigner.Public() // public key for creating server certificate
	if keyDER, err = serverKey.DER(); err != nil {
		t.Fatal(err)
	}
	writeFile(keyFilename, keyDER, ownerRW, t.Logf)

	t.Log("Creating server certificate")
	template = x509.Certificate{
		IPAddresses: []net.IP{IPv4loopback, net.IPv6loopback},
	}
	parlca.EnsureServer(&template)
	if certDER, err = caCert.Sign(&template, serverPublic); err != nil {
		t.Errorf("signing server certificate: %+v", err)
		return
	}
	writeFile(certFilename, certDER, ownerRW, t.Logf)

	// Listen() TLS()
	var https *Https = NewHttps(nearSocket, network, certDER, serverSigner)

	handler = newShandler()
	https.HandleFunc(URIPattern, handler.Handle)
	defer https.Shutdown()

	t.Log("invoking Listen")
	go errChListener(https.Listen(), goResult)

	t.Log("waiting for ListenAwaitable")
	<-https.ListenAwaitable.Ch()
	if !https.Near.IsValid() {
		t.Fatalf("FATAL: https.Near invalid")
	}
	near = https.Near.String()
	t.Logf("Near: %s", near)

	t.Log("issuing http.GET")
	resp, err = pnet.Get(protocol+near, pnet.NewTLSConfig(caX509), nil)
	// macOS does accept self-signed certificate
	//resp, err = http.Get(protocol + near)
	if resp != nil {
		statusCode = resp.StatusCode
		respS = fmt.Sprintf("status code: %d", statusCode)
	} else {
		respS = "resp nil"
	}

	t.Logf("%s err: %s", respS, perrors.Short(err))
	if err != nil {
		t.Errorf("http.Get err %s", perrors.Short(err))
	}
	// status code should be 204
	if resp != nil {
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("http.Get status code %d exp %d", resp.StatusCode, http.StatusNoContent)
		}
		if e := resp.Body.Close(); e != nil {
			panic(e)
		}
	}

	// handle count should be 1
	if c := int(handler.Rqs.Load()); c != 1 {
		t.Errorf("bad handle count: %d exp 1", c)
	}

	t.Logf("Shutting down server")
	https.Shutdown()

	// wait for error reader to exit
	goResult.ReceiveError(nil)

	if !https.EndListenAwaitable.IsClosed() {
		t.Error("EndListenAwaitable not closed")
	}
}

// writeFile writes byts to file panic on error
func writeFile(filename string, byts []byte, mode fs.FileMode, logf parl.PrintfFunc) {
	logf("Writing: %s", filename)
	if err := os.WriteFile(filename, byts, mode); err != nil {
		panic(perrors.Errorf("os.WriteFile: '%w'", err))
	}
}
