/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phttp

import (
	"log"
	"net/http"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/plog/plib"
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
	var writer = plib.NewWriter(logFunc)

	// [http.Server.ErrorLog] requires pointer to
	// concrete type [log.Logger]
	//	- has private fields that must be initialized
	//	- this costs allocation
	//	- Server object must be heap-allocated, too
	errorLog = log.New(writer, logPrefix, logFlags)

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

const (
	// prefix appears at the beginning of each generated log line
	logPrefix = ""
	// flags determines log-line formatting. 0 means transparent
	logFlags = 0
)
