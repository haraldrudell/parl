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

// MinimalRecovery handles process exit for main,
// in particular error message and non-zero status code for errors
//   - on success:
//   - –  a timestamped success message is printed to stderr
//   - — the function returns
//   - panics are not handled
//   - on error:
//   - if error contains a panic, error is printed with stack trace
//   - otherwise one-liner with location
//   - os.Exit is invoked with status code 1
//
// Usage:
//
//	main() {
//	  var err error
//	  defer mains.MinimalRecovery(&err)
func MinimalRecovery(errp *error) {

	var err error
	if errp != nil {
		err = *errp
	}
	var ts = time.Now().Format(parl.Rfc3339s)
	if err == nil {
		parl.Log("%s Completed successfully", ts)
		return // success return
	}
	var eStr string
	if isPanic, _, _, _ := perrors.IsPanic(err); isPanic {
		eStr = perrors.Long(err)
	} else {
		eStr = perrors.Short(err)
	}
	parl.Log("%s error: %s", ts, eStr)
	os.Exit(1) // failure process termination
}
