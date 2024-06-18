/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phttp

import (
	"context"
	"crypto/tls"
	"net/http"
)

// Get is a convenience method for [pnet.HttpClient.Get]
func Get(requestURL string, tlsConfig *tls.Config, ctx context.Context) (resp *http.Response, err error) {
	var req = NewRequest(requestURL, ctx, &err)
	if err != nil {
		return
	}
	resp, err = NewHttpClient(tlsConfig).Do(req)
	return
}
