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

// MinimalRecovery handles error process exit for the main function
//   - purpose is to avoid a silent zero status-code exit on error
//     and to print exactly when the process exited
//   - panics are not recovered
//   - on success: *errp == nil:
//   - –  a timestamped success message is printed to stderr
//   - — the function returns
//   - on error: *errp != nil:
//   - — if error is caused by panic, error is printed with stack trace
//   - — other errors are printed as a one-liner with code-location
//   - — os.Exit is invoked with status code 1
//
// Usage:
//
//	main() {
//	  var err error
//	  defer mains.MinimalRecovery(&err)
func MinimalRecovery(errp *error) {

	// process result
	var err error
	if errp != nil {
		err = *errp
	}
	// process exit timestamp
	var ts = time.Now().Format(parl.Rfc3339s)

	// good exit: return
	if err == nil {
		parl.Log("%s Completed successfully", ts)
		return // success return
	}

	// process exit with error
	//	- if panic: error with stack trace
	//	- other error: one-liner with code location
	var eStr string
	if isPanic, _, _, _ := perrors.IsPanic(err); isPanic {
		eStr = perrors.Long(err)
	} else {
		eStr = perrors.Short(err)
	}
	parl.Log("%s error: %s", ts, eStr)

	// exit code 1
	os.Exit(1)
}
