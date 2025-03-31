/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/pslices/pslib"

type AwaitableSliceState struct {
	Size, MaxRetainSize          int
	HasData                      uint32
	IsDataWaitActive             bool
	IsDataWaitClosed             bool
	IsCloseInvoked               bool
	IsClosed                     bool
	Primary, CachedInput, InList Metrics
	InQ                          []Metrics
	Head, CachedOutput, OutList  Metrics
	OutQ                         []Metrics
	ZeroOut                      pslib.ZeroOut
}

type Metrics struct {
	Length, Capacity int
}
