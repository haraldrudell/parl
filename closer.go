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
// Closer handles panics.
// if errp is non-nil, panic values updates it using errors.AppendError.
func Closer[T any](ch chan T, errp *error) {
	defer Recover(Annotation(), errp, NoOnError)

	close(ch)
}

// CloserSend is a deferrable function that closes a send-channel.
// CloserSend handles panics.
// if errp is non-nil, panic values updates it using errors.AppendError.
func CloserSend[T any](ch chan<- T, errp *error) {
	defer Recover(Annotation(), errp, NoOnError)

	close(ch)
}

// Close is a deferrable function that closes an io.Closer object.
// Close handles panics.
// if errp is non-nil, panic values updates it using errors.AppendError.
func Close(closable io.Closer, errp *error) {
	defer Recover(Annotation(), errp, NoOnError)

	if e := closable.Close(); e != nil {
		*errp = perrors.AppendError(*errp, e)
	}
}
