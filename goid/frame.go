/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package goid

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pruntime"
)

type Frame struct {
	pruntime.CodeLocation
	// args like "(1, 2, 3)"
	args string
}

var _ parl.Frame = &Frame{}

func (f *Frame) Loc() (location *pruntime.CodeLocation) {
	return &f.CodeLocation
}

func (f *Frame) Args() (args string) {
	return f.args
}
