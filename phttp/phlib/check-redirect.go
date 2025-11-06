/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phlib

import "net/http"

// CheckRedirect replaces default policy 10 consecutive requests
func CheckRedirect(req *http.Request, via []*http.Request) (err error) { return }

// NoRedirects does not follow redirects and returns eg. 308
func NoRedirects(req *http.Request, via []*http.Request) (err error) {
	// Return an error to prevent following redirects
	return http.ErrUseLastResponse
}
