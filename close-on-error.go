// © 2026–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License

package parl

import (
	"io"

	"github.com/haraldrudell/parl/perrors"
)

func CloseOnError(closer io.Closer, errp *error) {
	var err = *errp
	if err == nil {
		return
	}

	if e := closer.Close(); perrors.IsPF(&e, "Close %w", e) {
		*errp = perrors.AppendError(*errp, e)
	}
}
