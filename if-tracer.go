/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "time"

type Tracer interface {
	Assign(threadID, task string) (tracer Tracer)
	Record(threadID, text string) (tracer Tracer)
	Records(clear bool) (records map[string][]Record)
}

type Record interface {
	Values() (at time.Time, text string)
}