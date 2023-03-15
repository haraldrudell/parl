/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"os"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// MinimalRecovery prints error with location to os.Stderr and exits with status code 1
// if errp points to non-nil error.gp-timestamp
//
// Usage:
//
//	main() {
//	  var err error
//	  defer mains.MinimalRecovery(&err)
func MinimalRecovery(errp *error) {
	var ts = time.Now().Format(parl.Rfc3339s)
	if errp == nil || *errp == nil {
		parl.Log("%s Completed successfully", ts)
		return // no error return
	}
	parl.Log("%s error: %s", ts, perrors.Short(*errp))
	os.Exit(1)
}
