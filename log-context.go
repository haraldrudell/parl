/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "context"

const (
	// LogObject context key for log object in valuecontext.go
	LogObject = "parl.Log"
)

// GetLogFromContext obtains a log possibly from context
func GetLogFromContext(ctx context.Context) (log *LogInstance) {
	log = ctx.Value(LogObject).(*LogInstance)
	if log == nil {
		log = NewLog()
	}
	return
}

// GetLogAndContext gets a context and a log
func GetLogAndContext(debug bool, verbose bool, silent bool) (ctx context.Context, log *LogInstance) {
	ct1 := NewContext()
	log = NewLog()
	if debug && !silent {
		log.SetDebug(true)
	}
	/*
		TODO
		 else if verbose {
			log.SetVerbose()
		}
	*/
	StoreInContext(ct1, LogObject, log)
	ctx = ct1
	return
}
