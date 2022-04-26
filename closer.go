/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"io"

	"github.com/haraldrudell/parl/perrors"
)

// Closer is a deferrable function that closes a channel recovering
// from panic
func Closer[T any](ch chan T, errp *error) {
	defer Recover(Annotation(), errp, NoOnError)

	close(ch)
}

func CloserSend[T any](ch chan<- T, errp *error) {
	defer Recover(Annotation(), errp, NoOnError)

	close(ch)
}

func Close(closable io.Closer, errp *error) {
	defer Recover(Annotation(), errp, NoOnError)

	if e := closable.Close(); e != nil {
		*errp = perrors.AppendError(*errp, e)
	}
}
