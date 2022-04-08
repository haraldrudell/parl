/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"time"

	"github.com/haraldrudell/parl/goid"
)

type Tracer interface {
	Assign(threadID goid.ThreadID, task TracerTaskID) (tracer Tracer)
	Record(threadID goid.ThreadID, text string) (tracer Tracer)
	Records(clear bool) (records map[TracerTaskID][]TracerRecord)
}

type TracerTaskID string

type TracerRecord interface {
	Values() (at time.Time, text string)
}

type TracerFactory interface {
	NewTracer() (tracer Tracer)
}
