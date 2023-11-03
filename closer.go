/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"io"

	"github.com/haraldrudell/parl/perrors"
)

// Closer is a deferrable function that closes a channel.
//   - if errp is non-nil, panic values updates it using errors.AppendError.
//   - Closer is thread-safe, panic-free and deferrable
func Closer[T any](ch chan T, errp *error) {
	defer PanicToErr(errp)

	close(ch)
}

// CloserSend is a deferrable function that closes a send-channel.
//   - if errp is non-nil, panic values updates it using errors.AppendError.
//   - CloserSend is thread-safe, panic-free and deferrable
func CloserSend[T any](ch chan<- T, errp *error) {
	defer PanicToErr(errp)

	close(ch)
}

// Close closes an io.Closer object.
//   - if errp is non-nil, panic values updates it using errors.AppendError.
//   - Close is thread-safe, panic-free and deferrable
//   - type Closer interface { Close() error }
func Close(closable io.Closer, errp *error) {
	var err error
	defer RecoverErr(func() DA { return A() }, errp)

	if err = closable.Close(); perrors.IsPF(&err, "%w", err) {
		*errp = perrors.AppendError(*errp, err)
	}
}
