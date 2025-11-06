/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phttp

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/haraldrudell/parl/phttp/phlib"
)

// Get is a convenience method for [pnet.HttpClient.Get]
//   - Get offers over [http.Get]:
//   - — use of self-signed or client certificates
//   - — request-context
//   - — errors with stack trace
//   - Get is intended for one-off execution in tests,
//     Get creates a client object on every invocation
//   - for repeated Get, use [HttpClient.Do]
func Get(requestURL string, tlsConfig *tls.Config, ctx context.Context) (resp *http.Response, err error) {
	var req = NewRequest(requestURL, ctx, &err)
	if err != nil {
		return
	}
	// Do returns errors with stack-trace
	resp, err = NewHttpClient(tlsConfig).Do(req)
	return
}

func GetNoRedirects(requestURL string, tlsConfig *tls.Config, ctx context.Context) (resp *http.Response, err error) {
	var req = NewRequest(requestURL, ctx, &err)
	if err != nil {
		return
	}
	var c = NewHttpClient(tlsConfig)
	c.CheckRedirect = phlib.NoRedirects
	// Do returns errors with stack-trace
	resp, err = c.Do(req)
	return
}

// func http.Get(url string) (resp *http.Response, err error)
var _ = http.Get
