/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"unsafe"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/plog"
)

const (
	//   - [ZeroFillingShifter] provided as argument to [NewShifter] causes
	//     freed elements to be set to the T zero-value.
	//     This prevents temporary memory leaks when:
	//   - T contains pointers and
	//   - previously used T elements remain in the slice’s underlying array
	ZeroFillingShifter = true
)

// Shifter implements append filtered by a capacity check
//   - this avoids unnecessary slice re-allocations for when
//     a slice is used as a buffer:
//   - — items are sliced off at the beginning and
//   - — new items appended at the end
type Shifter[T any] struct {
	// Slice contains current elements
	Slice []T
	// initialSlice is the slice provided or from last re-allocating append
	initialSlice []T
	// if zeroFill [ZeroFillingShifter], slice element size in bytes
	elementSize int
}

// NewShifter returns a slice container providing an Append that filters
// unnecessary allocations
//   - if zeroFill is [ZeroFillingShifter] freed elements are set to zero-value
//   - — this prevents temporary memory leaks when:
//   - — elements contains pointers and
//   - — previously used elements remain in the slice’s underlying array
func NewShifter[T any](slice []T, zeroFill ...bool) (shifter *Shifter[T]) {
	var elementSize int
	if len(zeroFill) > 0 && zeroFill[0] {
		var s []T
		if cap(slice) >= 2 {
			s = slice[0:2]
		} else {
			s = make([]T, 2)
		}
		elementSize = //
			int(uintptr(unsafe.Pointer(&s[1]))) -
				int(uintptr(unsafe.Pointer(&s[0])))
	}
	return &Shifter[T]{Slice: slice, initialSlice: slice, elementSize: elementSize}
}

// Append avoids unnecessary allocations if capacity is sufficient
func (s *Shifter[T]) Append(items ...T) (slice []T) {
	if len(items) == 0 {
		return s.Slice // noop return
	}

	var requiredCapacity = len(s.Slice) + len(items)
	if requiredCapacity > cap(s.initialSlice) {

		// if insufficient capacity, use regular append
		s.Slice = append(s.Slice, items...)
	} else {

		// re-use initialSlice
		var xSlice = s.Slice
		s.Slice = append(append(s.initialSlice[:0], s.Slice...), items...)

		// do zero fill
		//	- when xSlice was more off from initialSlice than len(items)
		if s.elementSize > 0 && len(xSlice) > 0 {
			s.zeroFill(s.Slice, xSlice)
		}
	}
	s.initialSlice = s.Slice
	slice = s.Slice

	return
}

// zeroFill fills unused elements of now with zero-value
//   - now and before are non-empty slices
//   - before is expected to be a slice-result of now
//   - the task is to possibly write zero-values to last elements of before
func (s *Shifter[T]) zeroFill(now, before []T) {
	plog.D("now len %d before len %d", len(now), len(before))
	// first element of before as index of now
	var beforeIndex int
	var beforep = int(uintptr(unsafe.Pointer(&before[0])))
	var nowp = int(uintptr(unsafe.Pointer(&now[0])))
	beforeIndex = (beforep - nowp) / s.elementSize

	// check beforeIndex
	var capacity = cap(now)
	plog.D("beforeIndex: %d cap %d", beforeIndex, capacity)
	if beforeIndex < 0 || beforeIndex >= capacity {
		panic(s.zeroFillError(now, before, capacity, beforeIndex))
	}

	// zero-fill end of before
	var firstUnused = len(now) - beforeIndex
	plog.D("firstUnused: %d", firstUnused)
	var t T
	for firstUnused < len(before) {
		before[firstUnused] = t
		firstUnused++
	}
}

// zeroFillError returns an actionable error for a software deficiency
func (s *Shifter[T]) zeroFillError(now, before []T, capacity, beforeIndex int) (err error) {
	var bSign string
	var b = beforeIndex
	if b < 0 {
		bSign = "-"
		b = -beforeIndex
	}
	err = perrors.ErrorfPF("bad zeroFill: now: len %d cap %d before: len %d cap %d index: %s0x%x",
		len(now), capacity,
		len(before), cap(before),
		bSign, b,
	)
	return
}
