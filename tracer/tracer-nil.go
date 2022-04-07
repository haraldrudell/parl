/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package tracer

import "github.com/haraldrudell/parl"

type tracerNil struct{}

func NewTracerNil() (tracer parl.Tracer) {
	return &tracerNil{}
}

func (tn *tracerNil) Assign(threadID, task string) (tracer parl.Tracer) { return tn }
func (tn *tracerNil) Record(threadID, text string) (tracer parl.Tracer) { return tn }
func (tn *tracerNil) Records(clear bool) (records map[string][]parl.Record) {
	return map[string][]parl.Record{}
}
