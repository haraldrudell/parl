/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phttp

import (
	"crypto/tls"
	"net/http"

	"github.com/haraldrudell/parl/perrors"
)

// NewTransport returns a transport using tlsConfig
//   - based on [http.DefaultTransport]
func NewTransport(tlsConfig *tls.Config) (httpTransport *http.Transport) {
	var defaultTransport *http.Transport
	var ok bool
	if defaultTransport, ok = http.DefaultTransport.(*http.Transport); !ok {
		panic(perrors.New("DefaultTransport not http.Transport type"))
	}
	httpTransport = defaultTransport.Clone()

	// ensure tlsConfig is used
	httpTransport.TLSClientConfig = tlsConfig
	// must be nil to use TLSClientConfig
	httpTransport.DialTLSContext = nil

	return
}
