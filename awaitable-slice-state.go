/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/pslices/pslib"

type AwaitableSliceState struct {

	// static state

	Size, MaxRetainSize     int
	SizeMax4KiB, IsLowAlloc bool
	ZeroOut                 pslib.ZeroOut

	// global state

	IsDataWaitActive bool
	IsDataWaitClosed bool
	IsCloseInvoked   bool
	IsClosed         bool
	IsLength         bool
	Length           int
	MaxLength        uint64

	// output state

	Head, CachedOutput, OutList Metrics
	OutQ                        []Metrics
	HasData                     uint32
	LastPrimaryLarge            bool

	// input state

	Primary, CachedInput, InList Metrics
	InQ                          []Metrics
	HasInput, HasList            bool
}

type Metrics struct {
	Length, Capacity int
}

// State returns internal state for debug purposes
//   - populateValues missing: values are not retrieved
//   - populateValues [ValuesYes]: every queue slice value is provided in values
//   - holds both locks: not for frequent use
func (s *AwaitableSlice[T]) State(populateValues ...ValuesFlag) (state AwaitableSliceState, values AwaitableSliceValues[T]) {
	defer s.outQ.lock.Lock().Unlock()
	defer s.outQ.InQ.lock.Lock().Unlock()

	state = AwaitableSliceState{
		Size:          s.outQ.InQ.Size.Load(),
		MaxRetainSize: s.outQ.InQ.MaxRetainSize.Load(),
		SizeMax4KiB:   s.outQ.sizeMax4KiB.Load(),
		IsLowAlloc:    s.outQ.InQ.IsLowAlloc.Load(),
		ZeroOut:       s.outQ.InQ.ZeroOut.Load(),

		IsDataWaitActive: s.dataWait.IsActive.Load(),
		IsCloseInvoked:   s.isCloseInvoked.Load(),

		HasInput:         s.outQ.InQ.HasInput.Load(),
		HasList:          s.outQ.InQ.HasList.Load(),
		HasData:          uint32(s.outQ.HasDataBits.bits.Load()),
		IsLength:         s.outQ.InQ.IsLength.Load(),
		Length:           s.outQ.InQ.Length.Load(),
		MaxLength:        s.outQ.InQ.MaxLength.Max1(),
		LastPrimaryLarge: s.outQ.lastPrimaryLarge,

		Head: Metrics{
			Length:   len(s.outQ.head),
			Capacity: cap(s.outQ.head),
		},
		CachedOutput: Metrics{
			Length:   len(s.outQ.cachedOutput),
			Capacity: cap(s.outQ.cachedOutput),
		},
		OutList: Metrics{
			Length:   len(s.outQ.InQ.list.sliceList),
			Capacity: cap(s.outQ.InQ.list.sliceList),
		},

		Primary: Metrics{
			Length:   len(s.outQ.InQ.primary),
			Capacity: cap(s.outQ.InQ.primary),
		},
		CachedInput: Metrics{
			Length:   len(s.outQ.InQ.cachedInput),
			Capacity: cap(s.outQ.InQ.cachedInput),
		},
		InList: Metrics{
			Length:   len(s.outQ.InQ.list.sliceList),
			Capacity: cap(s.outQ.InQ.list.sliceList),
		},
	}
	if state.IsDataWaitActive {
		state.IsDataWaitClosed = s.dataWait.Cyclic.IsClosed()
	}
	if state.IsCloseInvoked {
		state.IsClosed = s.isEmpty.IsClosed()
	}

	if x := state.InList.Length; x > 0 {
		state.InQ = make([]Metrics, x)
		for i := range x {
			var slicep = &s.outQ.InQ.list.sliceList[i]
			state.InQ[i].Length = len(*slicep)
			state.InQ[i].Capacity = cap(*slicep)
		}
	}

	if x := state.OutList.Length; x > 0 {
		state.OutQ = make([]Metrics, x)
		for i := range x {
			var slicep = &s.outQ.list.sliceList[i]
			state.OutQ[i].Length = len(*slicep)
			state.OutQ[i].Capacity = cap(*slicep)
		}
	}

	if len(populateValues) == 0 || populateValues[0] != ValuesYes {
		return
	}
	s.getValues(&values)

	return
}
