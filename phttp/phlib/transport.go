/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phlib

import (
	"crypto/tls"
	"net/http"

	"github.com/haraldrudell/parl/perrors"
)

// NewTransport returns a transport with specific TLS configuration
//   - NewTransport allows use of self-signed and client certificates
//   - there is a process-wide shared [http.DefaultTransport]
//   - shares
func NewTransport(tlsConfig *tls.Config) (httpTransport *http.Transport) {

	// clone default transport to httpTransport
	if defaultTransport, ok := http.DefaultTransport.(*http.Transport); !ok {
		panic(perrors.New("DefaultTransport not http.Transport type"))
	} else {
		// allocation
		httpTransport = defaultTransport.Clone()
	}

	// ensure tlsConfig is used
	httpTransport.TLSClientConfig = tlsConfig
	// must be nil to use TLSClientConfig
	httpTransport.DialTLSContext = nil

	return
}

// var http.DefaultTransport http.RoundTripper
var _ = http.DefaultTransport
