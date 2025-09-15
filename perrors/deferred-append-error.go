/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import "github.com/haraldrudell/parl/internal/cyclebreaker2"

// DeferredAppendError copies any error in errSource
// to errDest
//   - deferrable version of [AppendError]
//   - use case: single-threaded, deferred function
//     aggregating function-local error value to errp
func DeferredAppendError(errSource, errDest *error) {
	cyclebreaker2.NilPanic("errSource", errSource)
	cyclebreaker2.NilPanic("errDest", errDest)
	var err = *errSource
	if err == nil {
		return // no error
	}
	*errDest = AppendError(*errDest, err)
}
