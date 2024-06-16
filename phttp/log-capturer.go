/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phttp

import (
	"io"
	"log"
	"net/http"

	"github.com/haraldrudell/parl"
)

// http.Server is Go’s http server
//   - http.Server.Serve randomly logs to either:
//   - — standard error using [log.Printf] or
//   - — using [http.Server.ErrorLog]
var _ = http.Server{}.ErrorLog

// [http.Server.ErrorLog] is [log.Logger] that has
//   - new function [log.New]
var _ = log.New

type LogCapturer struct{ log parl.PrintfFunc }

func NewLogCapturer(logF parl.PrintfFunc) (errorLog *log.Logger) {
	// c is [io.Writer] printing transparently to logF
	var c = LogCapturer{log: logF}
	// prefix appears at the beginning of each generated log line
	var prefix = ""
	// glags deals with log-line formatting. 0 means trasparent
	var flag = 0
	errorLog = log.New(&c, prefix, flag)
	return
}

var _ io.Writer

func (c *LogCapturer) Write(p []byte) (n int, err error) {
	n = len(p)
	c.log(string(p))
	return
}
