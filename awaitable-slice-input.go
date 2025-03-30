/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"

	"github.com/haraldrudell/parl/pslices/pslib"
)

// inputQueue contains enqueueing lock and structures
type inputQueue[T any] struct {
	// queue provides: lock sliceList sliceList0
	list[T]
	// primary is primary input-queue value-slice
	//	- not a slice-away slice
	//	- only appended to when s.queue.sliceList is empty
	//	- may be empty while queue.sliceList is non-empty
	primary []T
	// cachedInput is a pre-allocated value-slice behind inputQueue lock
	//	- used for primary or value-slice in sliceList
	//	- copied from cachedOutput while holding both locks
	//	- used by Send
	cachedInput []T
	// HasInput is true if inputQueue cachedInput is allocated
	HasInput atomic.Bool
	// HasList is true if inputQueue sliceList is allocated
	HasList atomic.Bool
	// size is allocation size for new value slices,
	// set by [AwaitableSlice.SetSize]
	//	- effective slice allocation capacity:
	//	- if size is unset or SetSize less than 1:
	//		larger of 4 KiB and 10 elements
	size Atomic64[int]
	// maxRetainSize is the longest slice capacity that will be reused
	//	- not retaining large slices avoids temporary memory leaks
	//	- reusing slices of reasonable size reduces allocations
	//	- effective value depends on size set by [AwaitableSlice.SetSize]
	//	- if size is unset or less than 100, use maxForPrealloc: 100
	//	- otherwise, use value provided to [AwaitableSlice.SetSize]
	maxRetainSize Atomic64[int]
	ZeroOut       Atomic32[pslib.ZeroOut]
}

func (i *inputQueue[T]) isEmptyPrimary() (isEmpty bool) {
	return len(i.primary) == 0
}

func (i *inputQueue[T]) getPrimarySlice() (primary []T) { return i.primary }

func (i *inputQueue[T]) send(value T) {

	// create primary
	if i.primary == nil {
		i.primary = i.createInputSlice(value)
		return // allocated primary input queue return
	}
	// primary is allocated

	// append single value to primary if sliceList is empty and value fits
	if i.isEmptyList() {
		if len(i.primary) < cap(i.primary) || cap(i.primary) < maxAppendValueSliceCapacity {
			i.enqueueInPrimary(value)
			return // appended value return
		}
	}
	// add value to sliceList

	// create new sliceList
	if !i.HasList.Load() {
		var valueSlice []T
		if cap(i.primary) >= maxAppendValueSliceCapacity {
			// large slice
			valueSlice = make([]T, 1, cap(i.primary))
			valueSlice[0] = value
		} else {
			valueSlice = i.createInputSlice(value)
		}
		var sliceList = i.makeSliceList(valueSlice)
		i.swapList(sliceList)
		i.HasList.Store(true)
		return // created new sliceList return
	}

	// last slice is known to be non-empty
	var length, capacity = i.getLastSliceMetrics()
	if length < capacity || capacity < maxAppendValueSliceCapacity {
		// append to ending local slice
		i.enqueueInLastSlice(value)
		return // appended to last slice return
	}

	// append new slice
	var valueSlice []T
	if capacity >= maxAppendValueSliceCapacity {
		// large slice
		valueSlice = make([]T, 1, capacity)
		valueSlice[0] = value
	} else {
		valueSlice = i.createInputSlice(value)
	}
	i.enqueueInList(valueSlice)
}

// getPrimaryMetrics returns capacity of available inputQueue value-slice
//   - 0: no allocation is available
func (i *inputQueue[T]) getPrimaryMetrics() (length, capacity int) {

	// try i.q
	if capacity = cap(i.primary); capacity > 0 {
		length = len(i.primary)
		return
	}

	// try cachedInput
	capacity = cap(i.cachedInput)
	length = len(i.cachedInput)
	return
}

// swapPrimary gets primary input queue and its ownership
func (i *inputQueue[T]) swapPrimary(set ...[]T) (q []T) {

	// set case
	if len(set) > 0 {
		i.primary = set[0]
		return
	}

	//get case
	q = i.primary
	i.primary = nil

	return
}

// createInputSlice creates a value slice possibly using cachedInput
//   - values are copied
func (i *inputQueue[T]) createInputSlice(values ...T) (valueSlice []T) {
	if i.HasInput.Load() && len(values) <= cap(i.cachedInput) {
		valueSlice = i.cachedInput[:len(values)]
		i.cachedInput = nil
		i.HasInput.Store(false)
		copy(valueSlice, values)
		return
	}
	valueSlice = i.makeValueSlice(values...)
	return // allocated primary input queue return
}

// enqueueInPrimary stores value in q
func (i *inputQueue[T]) enqueueInPrimary(values ...T) {

	// use cached input if:
	//	- q is not alocated
	//	- values fit cachedInput
	if i.primary == nil && len(values) <= cap(i.cachedInput) {
		i.primary = i.cachedInput
		i.cachedInput = nil
	}

	// if values fit q, copy
	if n := len(i.primary) + len(values); n <= cap(i.primary) {
		var n0 = len(i.primary)
		i.primary = i.primary[:n]
		copy(i.primary[n0:], values)
		return // copy return
	}

	// realloc q
	i.primary = append(i.primary, values...)
}

// setCachedInput sets or gets cachedInput
func (i *inputQueue[T]) setCachedInput(slice ...[]T) {

	i.cachedInput = slice[0]
	var has = i.cachedInput != nil
	if has != i.HasInput.Load() {
		i.HasInput.Store(has)
	}
}

// setAndDiscardPrimary sets new primary and
// tries to retain primary as inputCached
func (i *inputQueue[T]) setAndDiscardPrimary(slice []T) {
	if i.cachedInput == nil && cap(i.primary) > 0 && cap(i.primary) <= i.maxRetainSize.Load() {
		if len(i.primary) > 0 {
			if i.ZeroOut.Load() == DoZeroOut {
				clear(i.primary)
			}
			i.cachedInput = i.primary[:0]
			return
		}
		i.cachedInput = i.primary
		i.HasInput.Store(true)
	}
	i.primary = slice
}

// makeValueSlice returns an allocated slice capacity SetSize default 10
// empty slice returns a new slice of length 0 and configured capacity
//   - value: if present, is added to the new slice
//   - newSlice an allocated slice, length 0 or 1 if value present
func (s *inputQueue[T]) makeValueSlice(values ...T) (newSlice []T) {

	// ensure size sizeMax are initialized
	var size = max(s.size.Load(), len(values))

	// create slice optionally with value
	newSlice = make([]T, len(values), size)
	if len(values) == 0 {
		return
	} else if len(values) == 1 {
		newSlice[0] = values[0]
		return
	}
	copy(newSlice, values)
	return
}

// makeSliceList makes slice list for inQ or outQ
//   - slice: optional slice as first element
func (s *inputQueue[T]) makeSliceList(slice ...[]T) (newSliceList [][]T) {
	var length int
	if len(slice) > 0 {
		length = 1
	}
	newSliceList = make([][]T, length, sliceListSize)
	if len(slice) > 0 {
		newSliceList[0] = slice[0]
	}
	return
}
