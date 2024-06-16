/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phttp

import (
	"context"
	"crypto"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"testing"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/parlca"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pnet"
)

func TestHttps(t *testing.T) {
	//t.Error("Logging on")
	const (
		// GET request protocol: “https://”
		protocol = "https://"
		// server URI: “/” matches everything
		URIPattern = "/"
		// ca common name, will default to {host}ca
		canonicalName       = ""
		httpShutdownTimeout = 5 * time.Second
		expRequestCount     = 1
	)
	var (
		// listening socket is IPv4 localhost ephemeral port
		socketAddress = pnet.NewSocketAddressLiteral(
			pnet.NetworkTCP4,
			netip.MustParseAddrPort("127.0.0.1:0"),
		)
		// IP address for certificate
		IPv4loopback = net.IPv4(127, 0, 0, 1)
	)

	var (
		err error
		// caCert is binary private key and binary DER ASN.1 certificate
		caCert parl.CertificateAuthority
		// ca certificate in usable [x509.Certificate] format
		caX509 *x509.Certificate
		// serverSigner is binary and [crypto.Signer] used to run the server
		serverSigner parl.PrivateKey
		// public key for creating server certificate
		serverPublic   crypto.PublicKey
		template       x509.Certificate
		certDER        parl.CertificateDer
		requestCounter *sHandler
		respS          string
		resp           *http.Response
		statusCode     int
		getURI         string
		ctx            = context.Background()
		shutdownCh     parl.AwaitableCh
		errIterator    parl.ErrsIter
		errList        []error
	)

	// methods to test:
	//	- HandleFunc() Listen() TLS() Serve()
	//	- Errs() ShutdownCh()
	//	- Shutdown() Shutdown2()
	var httpsServer *Https

	// create http Server
	// ensure credentials
	t.Log("Creating self-signed certificate authority")
	// caCert is binary private key and binary DER ASN.1 certificate
	if caCert, err = parlca.NewSelfSigned(canonicalName, x509.RSA); err != nil {
		// x509.RSA: “RSA”
		t.Fatalf("FAIL parlca.NewSelfSigned %s “%s”", x509.RSA, perrors.Short(err))
	}
	// expand certificate to [x509.Certificate[]
	if caX509, err = caCert.Check(); err != nil {
		t.Fatalf("FAIL: caCert.Check: %s", perrors.Short(err))
	}
	t.Log("Creating server private key")
	// serverSigner is binary and [crypto.Signer] used to run the server
	if serverSigner, err = parlca.NewEd25519(); err != nil {
		t.Fatalf("FAIL server parlca.NewEd25519: “%q”", err)
	}
	t.Log("Creating server certificate")
	// public key for creating server certificate
	serverPublic = serverSigner.Public()
	template = x509.Certificate{
		IPAddresses: []net.IP{IPv4loopback, net.IPv6loopback},
	}
	// certificate use is server authentication
	parlca.EnsureServer(&template)
	// have ca sign the certificate into binary DER ASN.1 form
	if certDER, err = caCert.Sign(&template, serverPublic); err != nil {
		t.Fatalf("FAIL signing server certificate: “%s”", err)
	}
	httpsServer = NewHttps(certDER, serverSigner)

	// add handler shared by all listeners counting requests
	requestCounter = newShandler()
	httpsServer.HandleFunc(URIPattern, requestCounter.Handle)

	// listen should trigger event
	t.Log("invoking Listen…")
	if nearAddrPort, listener, e := httpsServer.Listen(socketAddress); e == nil {
		go httpsServer.Serve(listener)
		t.Logf("near addr-port: %s", nearAddrPort)
		getURI = protocol + nearAddrPort.String()
	} else {
		t.Fatalf("Listen err “%s”", e)
	}

	// GET should succeed with status code 204
	t.Log("issuing http.GET…")
	resp, err = pnet.Get(getURI, pnet.NewTLSConfig(caX509), ctx)
	// macOS does accept self-signed certificate
	//resp, err = http.Get(protocol + near)
	respS = ""
	if resp != nil {
		statusCode = resp.StatusCode
		respS = fmt.Sprintf("status code: %d", statusCode)
	}
	if err != nil {
		respS += fmt.Sprintf("Get err “%s”", err)
	}
	t.Logf(respS)
	if err != nil {
		t.Errorf("FAIL http.Get err %s", perrors.Short(err))
	} else if resp == nil {
		t.Fatal("resp nil")
	}
	// GET should return status code 204
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("FAIL http.Get status code %d exp %d", resp.StatusCode, http.StatusNoContent)
	} else if e := resp.Body.Close(); e != nil {
		t.Fatalf("FAIL resp.Body.Close err %s", e)
	}

	// handle count should be 1
	if c := int(requestCounter.RequestCount.Load()); c != expRequestCount {
		t.Errorf("FAIL bad handle count: %d exp %d", c, expRequestCount)
	}

	// ShutdownCh should be non-nil untriggered
	shutdownCh = httpsServer.ShutdownCh()
	if shutdownCh == nil {
		t.Error("ShutdownCh nil")
	}
	select {
	case <-shutdownCh:
		t.Errorf("ShutdownCh triggered")
	default:
	}

	// on shutdown, endListen should trigger
	t.Logf("Shutting down server…")
	err = httpsServer.Shutdown()
	if err != nil {
		t.Errorf("Shutdown err “%s”", err)
	}

	// ShutdownCh should be triggered
	select {
	case <-shutdownCh:
	default:
		t.Errorf("ShutdownCh untriggered")
	}

	// Errs should be empty
	errIterator = httpsServer.Errs()
	errList = errIterator.Errors()
	if len(errList) > 0 {
		t.Errorf("Server had errors: %v", errList)
	}

	httpsServer.Shutdown2(&err)
	if err != nil {
		t.Errorf("Shutdown2 err “%s”", err)
	}
}

// sHandler counts incoming requests
type sHandler struct{ RequestCount parl.Atomic64[int] }

// newShandler returns a request counter
func newShandler() (s *sHandler) { return &sHandler{} }

// Handle is the http-server handler function
//   - return body-less status Code 204: no content
//   - counts requests
func (s *sHandler) Handle(w http.ResponseWriter, r *http.Request) {
	s.RequestCount.Add(1)
	w.WriteHeader(http.StatusNoContent)
}
