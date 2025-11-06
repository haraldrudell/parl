/*
© 2025–present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package phttp

import (
	"fmt"
	"net/http"
)

// StatusLine returns http 1.1 status line “HTTP/1.1 200 OK”
func StatusLine(statusCode int) (statusLine string) {

	var statusText = http.StatusText(statusCode)
	if statusText != "" {
		statusText = "\x20" + statusText
	}

	statusLine = fmt.Sprintf("%s %d%s",
		httpVersion,
		http.StatusOK,
		statusText,
	)

	return
}

const (
	httpVersion = "HTTP/1.1"
)
