/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/pslices/pslib"

type AwaitableSliceState struct {
	Size, MaxRetainSize     int
	SizeMax4KiB, IsLowAlloc bool
	ZeroOut                 pslib.ZeroOut

	IsDataWaitActive bool
	IsDataWaitClosed bool
	IsCloseInvoked   bool
	IsClosed         bool
	IsLength         bool
	Length           int
	MaxLength        uint64

	Head, CachedOutput, OutList Metrics
	OutQ                        []Metrics
	HasData                     uint32
	LastPrimaryLarge            bool

	Primary, CachedInput, InList Metrics
	InQ                          []Metrics
	HasInput, HasList            bool
}

type Metrics struct {
	Length, Capacity int
}
