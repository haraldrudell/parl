/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package tracer

import "github.com/haraldrudell/parl"

var TracerFactory parl.TracerFactory = &tracerFactory{}

type tracerFactory struct{}

func (tf *tracerFactory) NewTracer(use bool) (tracer parl.Tracer) {
	if use {
		tracer = NewTracer()
	} else {
		tracer = NewTracerNil()
	}
	return
}
