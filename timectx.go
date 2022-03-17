/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"time"
)

const (
	// TimePointer is context key for a pointer to a shared timestamp
	TimePointer = "parl.Time"
)

var defaultTime, _ = time.Parse(time.RFC3339, "2000-01-01T00:00:00Z")

// ContextTime obtains a shared timestamp
func ContextTime(ctx context.Context) (t time.Time, found bool) {
	var timePointer *time.Time
	contextValue := ctx.Value(TimePointer)
	if timePointer, found = contextValue.(*time.Time); found {
		t = *timePointer
	} else {
		t = defaultTime
	}
	return
}
