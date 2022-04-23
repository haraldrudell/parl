/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package tracer

import (
	"github.com/haraldrudell/parl"
)

type tracerNil struct{}

func NewTracerNil() (tracer parl.Tracer) {
	return &tracerNil{}
}

func (tn *tracerNil) AssignTaskToThread(threadID parl.ThreadID, task parl.TracerTaskID) (tracer parl.Tracer) {
	return tn
}
func (tn *tracerNil) RecordTaskEvent(threadID parl.ThreadID, text string) (tracer parl.Tracer) {
	return tn
}
func (tn *tracerNil) Records(clear bool) (records map[parl.TracerTaskID][]parl.TracerRecord) {
	return map[parl.TracerTaskID][]parl.TracerRecord{}
}
