/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlnet

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"

	"github.com/haraldrudell/parl/error116"
)

type HttpClient struct {
	http.Client
}

func Get(requestURL string, tlsConfig *tls.Config, ctx context.Context) (resp *http.Response, err error) {
	return NewHttpClient(tlsConfig).Get(requestURL, ctx)
}

func NewTLSConfig(cert *x509.Certificate) (tlsConfig *tls.Config) {
	certPool := x509.NewCertPool()
	certPool.AddCert(cert)
	return &tls.Config{RootCAs: certPool}
}

func NewHttpClient(tlsConfig *tls.Config) (httpClient *HttpClient) {
	return &HttpClient{Client: http.Client{
		Transport:     NewTransport(tlsConfig),
		CheckRedirect: CheckRedirect,
		Jar:           nil,
		Timeout:       0,
	}}
}

func (ct *HttpClient) Get(requestURL string, ctx context.Context) (resp *http.Response, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var httpRequest *http.Request
	if httpRequest, err = http.NewRequestWithContext(ctx, "GET", requestURL, nil); err != nil {
		panic(error116.Errorf("http.NewRequestWithContext: '%w'", err))
	}
	return ct.Client.Do(httpRequest)
}

func CheckRedirect(req *http.Request, via []*http.Request) (err error) {
	return
}

func NewTransport(tlsConfig *tls.Config) (httpTransport *http.Transport) {
	var defaultTransport *http.Transport
	var ok bool
	if defaultTransport, ok = http.DefaultTransport.(*http.Transport); !ok {
		panic(error116.New("DefaultTransport not http.Transport type"))
	}
	httpTransport = defaultTransport.Clone()

	// ensure tlsConfig is used
	httpTransport.TLSClientConfig = tlsConfig
	httpTransport.DialTLSContext = nil
	return
}
