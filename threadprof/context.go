/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package ghandi interfaces Android devices
package threadprof

import (
	"context"

	"github.com/haraldrudell/parl"
)

const CounterKey = "progressorCounters"

func CountersInContext(ctx context.Context, counters Counters) (ctx2 context.Context) {
	if cs := contextCounters(ctx); cs != nil {
		return ctx // counters already present
	}
	return context.WithValue(ctx, CounterKey, counters)
}

func CountersFromContext(ctx context.Context) (counters Counters) {
	if counters = contextCounters(ctx); counters == nil {
		parl.Log("CONTEXTNOCOUNTERS")
		panic("NO OCUNTERS")
	}
	return
}

func contextCounters(ctx context.Context) (cs *CountersOn) {
	cs, _ = ctx.Value(CounterKey).(*CountersOn)
	return
}
