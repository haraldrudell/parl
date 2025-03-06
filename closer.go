/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"io"

	"github.com/haraldrudell/parl/perrors"
)

// Closer is a deferrable generic function closing a channel
//   - ch: channel to close, closed or nil returns error
//   - — [CloserSend] is same for send-only channel
//   - errp non-nil: receives any panic using [perrors.AppendError]
//   - thread-safe panic-free deferrable
//   - if a thread is blocked in channel send for ch, close causes data race
func Closer[T any](ch chan T, errp *error) {
	defer RecoverErr(func() DA { return A() }, errp)

	close(ch)
}

// CloserSend is a deferrable function closing a send-only channel
//   - ch: channel to close, closed or nil returns error
//   - — similar to [Closer] but for send-only channel
//   - errp non-nil: receives any panic using [perrors.AppendError]
//   - thread-safe panic-free deferrable
//   - if a thread is blocked in channel send for ch, close causes data race
func CloserSend[T any](ch chan<- T, errp *error) {
	defer RecoverErr(func() DA { return A() }, errp)

	close(ch)
}

// Close closes an [io.Closer] as deferred function
//   - closable: type Closer interface { Close() error }
//   - — nil returns error. Already closed may return error
//   - errp: receives errors or panic via [perrors.AppendError]
//   - thread-safe panic-free deferrable
func Close(closable io.Closer, errp *error) {
	defer RecoverErr(func() DA { return A() }, errp)

	var err error
	if err = closable.Close(); err == nil {
		return // successful close
	}
	// Close returned error

	// ensure err has stack trace
	if !perrors.HasStack(err) {
		// the stack should begin with caller of parl.Close:
		err = perrors.Stackn(err, 1)
	}

	*errp = perrors.AppendError(*errp, err)
}
