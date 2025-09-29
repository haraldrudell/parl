/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phlib

import "net/http"

// CheckRedirect replaces default policy 10 consecutive requests
func CheckRedirect(req *http.Request, via []*http.Request) (err error) { return }
