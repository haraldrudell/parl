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

// HttpClient implements http GET for specific TLS configuration
//   - [http.DefaultClient] uses [http.DefaultTransport]
//   - — has no timeouts
//   - — cannot use client certificate
type HttpClient struct{ http.Client }

// NewHttpClient returns a client based on [http.Client] and tlsConfig
//   - HttpClient features custom transport to facilitate TLS setting
//     such as accepting server self-signed certificates or client certificate
//   - tlsConfig: optional configuration for https.
//     May be obtained from [pnet.NewTLSConfig]
//   - [phttp.NewRequest] obtains a request object that can be modified
//   - — method, request headers, cookies, body
//   - [HttpClient.Do] issues request
func NewHttpClient(tlsConfig *tls.Config) (httpClient *HttpClient) {
	return &HttpClient{Client: http.Client{
		Transport:     NewTransport(tlsConfig),
		CheckRedirect: CheckRedirect,
		Jar:           nil,
		Timeout:       0,
	}}
}

// func http.Get(url string) (resp *http.Response, err error)
var _ = http.Get

// used by [http.Get]
var _ = http.DefaultClient

// [http.Get] transport
var _ = http.DefaultTransport

// Request struct
var _ http.Request

func (c *HttpClient) Do(req *http.Request) (resp *http.Response, err error) {
	if resp, err = c.Client.Do(req); err != nil {
		err = perrors.ErrorfPF("http.Client.Do %w", err)
	}
	return
}
