/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phttp

import (
	"context"
	"net/http"

	"github.com/haraldrudell/parl/perrors"
)

func NewRequest(requestURL string, ctx context.Context, errp *error) (req *http.Request) {
	var err error
	if req, err = http.NewRequestWithContext(ctx, GetMethod, requestURL, nil); err != nil {
		*errp = perrors.AppendError(*errp,
			perrors.Errorf("http.NewRequestWithContext: “%w” requestURL: %q", err, requestURL),
		)
	}
	return
}
