/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package tracer

import (
	"context"

	"github.com/haraldrudell/parl"
)

const TracerKey = "tracerTracer"

func TracerInContext(ctx context.Context, tracer parl.Tracer) (ctx2 context.Context) {
	if cs := contextTracer(ctx); cs != nil {
		return ctx // already present
	}
	return context.WithValue(ctx, TracerKey, tracer)
}

func TracerFromContext(ctx context.Context) (tracer parl.Tracer) {
	if tracer = contextTracer(ctx); tracer == nil {
		parl.Log("CONTEXTNOCOUNTERS")
		panic("NO OCUNTERS")
	}
	return
}

func contextTracer(ctx context.Context) (cs parl.Tracer) {
	cs, _ = ctx.Value(TracerKey).(parl.Tracer)
	return
}
