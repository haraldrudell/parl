/*
© 2025–present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package phttp_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"net/netip"
	"os"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/parlca"
	"github.com/haraldrudell/parl/parlca/calib"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/phttp"
	"github.com/haraldrudell/parl/pnet"
	"github.com/haraldrudell/parl/pos"
)

func TestServerHttps(t *testing.T) {
	// t.Error("Logging on")
	const (
		validateCredentials = false
		writeCredentials    = false
		generateCredentials = false
	)

	const (
		// localhost high ephemeral port
		// serverSocketAddr = "localhost:0"
		expStatusCode = http.StatusNotFound
	)

	var (
		cert            parl.Certificate
		privateKey      parl.PrivateKey
		ctx             = context.Background()
		err             error
		tlsListener     net.Listener
		isUp            = make(chan struct{})
		isDone          = make(chan struct{})
		requestAddr     string
		resp            *http.Response
		x509Certfiicate *x509.Certificate
		tlsConfig       *tls.Config
	)

	cert, privateKey = handleCredentials(generateCredentials, validateCredentials, writeCredentials, t)
	x509Certfiicate, err = cert.ParseCertificate()
	if err != nil {
		t.Fatalf("FAIL ParseCertificate err ‘%s’", perrors.Short(err))
	}

	// handler null will return status code 404
	var handler http.Handler

	// Listen() PoolServe() Close()
	var cDavServer *phttp.Server = phttp.NewServerHandler(handler, t.Logf)

	// listening to https socket should succeed
	var nearSocket netip.AddrPort
	tlsListener, nearSocket, err = cDavServer.Listen(cert.DER(), privateKey)
	if err != nil {
		t.Fatalf("FAIL Listen err ‘%s’", perrors.Short(err))
	}

	// print near-socket address
	t.Logf("listen address: %s", nearSocket)

	// listen in separate thread
	go listenThread(cDavServer, tlsListener, t, isUp, isDone)
	<-isUp

	// nearSocket may be “::” “0.0.0.0” but must be “::1” or “127.0.0.1”
	var requestSocket netip.AddrPort
	if a := nearSocket.Addr(); !a.IsUnspecified() {
		requestSocket = nearSocket
	} else if a.Is4() {
		requestSocket = netip.AddrPortFrom(pnet.LoopbackIpv4, nearSocket.Port())
	} else {
		requestSocket = netip.AddrPortFrom(pnet.LoopbackIpv6, nearSocket.Port())
	}

	// issue https request to listening server
	requestAddr = "https://" + requestSocket.String()
	t.Logf("issuing GET request to %s", requestAddr)
	tlsConfig = pnet.NewTLSConfig(x509Certfiicate)
	resp, err = phttp.Get(requestAddr, tlsConfig, ctx)
	if err != nil {
		t.Errorf("FAIL http.Get err ‘%s’", perrors.Short(err))
	} else {
		// a GET reponse was received

		if resp.StatusCode != expStatusCode {
			t.Errorf("phttp.Get status code: %d exp %d", resp.StatusCode, expStatusCode)
		} else {
			t.Logf("OK GET response status code: %d", resp.StatusCode)
		}
		err = resp.Body.Close()
		if err != nil {
			t.Errorf("FAIL Body.Close err ‘%s’", perrors.Short(err))
		}
	}

	// close server: garceful close
	err = cDavServer.Close()
	if err != nil {
		t.Errorf("FAIL Close err ‘%s’", perrors.Short(err))
	} else {
		t.Logf("server.Close succeeded")
	}

	// await listenThread exit
	<-isDone
}

// listenThread runs listening server
func listenThread(cDavServer *phttp.Server, tlsListener net.Listener, t *testing.T, isUp, isDone chan struct{}) {
	defer close(isDone)
	close(isUp)

	var err = cDavServer.Serve(tlsListener)
	if err != nil {
		t.Errorf("FAIL listenThread func PoolServe err ‘%s’", perrors.Short(err))
	}
}

// handleCredentials obtains cert and privateKey by
// generating, validating and writing credentials
func handleCredentials(
	generateCredentials, validateCredentials, writeCredentials bool,
	t *testing.T,
) (cert parl.Certificate, privateKey parl.PrivateKey) {
	var err error

	if generateCredentials {
		cert, privateKey = parlca.CreateRSA()
		t.Logf("%s", cert.PEM())
		t.Logf("%s", privateKey.PEMe())
	} else {
		cert, _, _, err = parlca.ParsePem([]byte(calib.CertLocalhost))
		if err != nil {
			t.Fatalf("FAIL serverCred cert err ‘%s’", perrors.Short(err))
		}
		_, privateKey, _, err = parlca.ParsePem([]byte(calib.KeyLocalhost))
		if err != nil {
			t.Fatalf("FAIL serverCred key err ‘%s’", perrors.Short(err))
		}
	}

	// check key and certificate
	if validateCredentials {
		err = privateKey.Validate()
		if err != nil {
			t.Fatalf("FAIL bad private key err ‘%s’", perrors.Short(err))
		}
		var x509Certificate *x509.Certificate
		x509Certificate, err = cert.ParseCertificate()
		if err != nil {
			t.Fatalf("FAIL bad certificate err ‘%s’", perrors.Short(err))
		}
		_ = x509Certificate
	}

	// write certificate to file system
	if writeCredentials {
		// gapi/gonty/cert-TestCDavServerHttps.pem
		// openssl x509 -noout -text -in gapi/gonty/cert-TestCDavServerHttps.pem | less
		err = os.WriteFile("cert-TestCDavServerHttps.pem", cert.PEM(), pos.PermUserReadWrite)
		if err != nil {
			t.Fatalf("FAIL WriteFile err ‘%s’", perrors.Short(err))
		}
	}

	return
}
