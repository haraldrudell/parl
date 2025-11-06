/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phttp

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/url"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/phttp/phlib"
)

// HttpClient implements http GET for specific TLS configuration
//   - HttpClient provides Do method with:
//   - — custom TLS transport for self-signed or client certificates
//   - — errors with stack trace
//   - — use of context in request value
//   - Do is used by [phttp.Get] and other code
//   - there is no interface for [http.Client], it is struct, so HttpClient is struct, too
//   - [http.DefaultClient] uses [http.DefaultTransport]
//   - — has no timeouts
//   - — cannot use client certificate
type HttpClient struct{ http.Client }

// NewHttpClient returns a client based on [http.Client] and tlsConfig
//   - HttpClient provides Do method with:
//   - — custom TLS transport for self-signed or client certificates
//   - — errors with stack trace
//   - — use of context in request value
//   - tlsConfig: optional configuration for https, may be nil.
//     May be obtained from [pnet.NewTLSConfig]
//   - [phttp.NewRequest] obtains a request object that can be modified
//   - — method, request headers, cookies, body
//   - [HttpClient.Do] issues request
func NewHttpClient(tlsConfig *tls.Config) (httpClient *HttpClient) {
	return &HttpClient{Client: http.Client{
		Transport: phlib.NewTransport(tlsConfig),
		//CheckRedirect: phlib.CheckRedirect,
		Jar:     nil,
		Timeout: 0,
	}}
}

// Do sends an HTTP request and returns an HTTP response, following
// policy (such as redirects, cookies, auth) as configured on the
// client.
func (c *HttpClient) Do(req *http.Request) (resp *http.Response, err error) {
	if resp, err = c.Client.Do(req); err != nil {
		err = perrors.ErrorfPF("http.Client.Do %w", err)
	}
	return
}

// Get issues a GET to the specified URL
func (c *HttpClient) Get(url string) (resp *http.Response, err error) {
	if resp, err = c.Client.Get(url); err != nil {
		err = perrors.ErrorfPF("http.Client.Get %w", err)
	}
	return
}

// Head issues a HEAD to the specified URL
func (c *HttpClient) Head(url string) (resp *http.Response, err error) {
	if resp, err = c.Client.Head(url); err != nil {
		err = perrors.ErrorfPF("http.Client.Head %w", err)
	}
	return
}

// Post issues a POST to the specified URL
func (c *HttpClient) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	if resp, err = c.Client.Post(url, contentType, body); err != nil {
		err = perrors.ErrorfPF("http.Client.Post %w", err)
	}
	return
}

// PostForm issues a POST to the specified URL, with data's keys and values URL-encoded as the request body
func (c *HttpClient) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	if resp, err = c.Client.PostForm(url, data); err != nil {
		err = perrors.ErrorfPF("http.Client.PostForm %w", err)
	}
	return
}

// Get is the convenience function provided by the http package for http GET
// func http.Get(url string) (resp *http.Response, err error)
var _ = http.Get

// [http.Get] delegates to [http.DefaultClient] used by [http.Get]
var _ = http.DefaultClient

// the type of [http.DefaultClient] is struct [http.Client]
//   - methods: CloseIdleConnections() Do() Get() Head() Post() PostForm()
var _ http.Client

// func (c *http.Client) Do(req *http.Request) (*http.Response, error)
var _ = (&http.Client{}).Do

// [http.Get] transport
var _ = http.DefaultTransport

// Request struct
var _ http.Request
