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

// NewErrorLog can be assigned to [http.Server.ErrorLog]
// to forward any logs to the log function
//   - log: delegated-to log writer
//   - —
//   - http.Server logs occasionally to standard error
//   - by assigning the ErrorLog field those logs can be
//     channeled
//   - http.Server methods requires Server struct to be
//     heap allocated
//   - that means ErrorLog must be heap-allocated, too
func NewErrorLog(logFunc parl.PrintfFunc) (errorLog *log.Logger) {

	// log.New requires an [io.Writer]
	//	- this means heap allocation
	var writer = printFuncWriter{
		log: logFunc,
	}

	// [http.Server.ErrorLog] requires pointer to
	// concrete type [log.Logger]
	//	- has private fields that must be initialized
	//	- this costs allocation
	//	- Server object must be heap-allocated, too
	errorLog = log.New(&writer, logPrefix, logFlags)

	return
}

// http.Server is Go’s http server
//   - http.Server.Serve randomly logs to either:
//   - — standard error using [log.Printf] or
//   - — using [http.Server.ErrorLog]
//   - ErrorLog *log.Logger
var _ = (&http.Server{}).ErrorLog

// [http.Server.ErrorLog] is [log.Logger] that has
//   - new function [log.New]
//   - must be used to initialize the writer field
var _ = log.New

// printFuncWriter implements an [io.Writer] logging to log
type printFuncWriter struct {
	log parl.PrintfFunc
}

// printFuncWriter is [io.Writer]
var _ io.Writer = &printFuncWriter{}

// Write converts write of bytes to log of string
func (c *printFuncWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	c.log(string(p))
	return
}

const (
	// prefix appears at the beginning of each generated log line
	logPrefix = ""
	// flags determines log-line formatting. 0 means transparent
	logFlags = 0
)
