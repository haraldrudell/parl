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
	// HasInput is true if inputQueue cachedInput is allocated
	//	- written behind InQ lock
	HasInput atomic.Bool
	// HasList is true if inputQueue sliceList is allocated
	//	- written behind InQ lock
	HasList atomic.Bool
	// Size is allocation Size for new value slices,
	// set by [AwaitableSlice.SetSize]
	//	- effective slice allocation capacity:
	//	- if Size is unset or SetSize less than 1:
	//		larger of 4 KiB and 10 elements
	Size Atomic64[int]
	// MaxRetainSize is the longest slice capacity that will be reused
	//	- not retaining large slices avoids temporary memory leaks
	//	- reusing slices of reasonable size reduces allocations
	//	- effective value depends on size set by [AwaitableSlice.SetSize]
	//	- if size is unset or less than 100, use maxForPrealloc: 100
	//	- otherwise, use value provided to [AwaitableSlice.SetSize]
	MaxRetainSize Atomic64[int]
	// ZeroOut controls zero-out of unused value-slice elements
	//	- T types like pointers or interface-values that may hold references
	//		to other memory are zeroed out
	//	- other types, eg. bytes, are not zeroed out
	ZeroOut Atomic32[pslib.ZeroOut]
	// IsLength is true if the queue tracks length and maxLength
	//	- written behind both InQ lock and outQ lock
	IsLength atomic.Bool
	// Length is current length of queue
	//	- valid when length is being tracked
	Length Atomic64[int]
	// MaxLength is the maximum length of the queue
	//	- valid when length is being tracked
	MaxLength AtomicMax[uint64]
	// IsLowAlloc disables pre-allocation
	//   - type T is error and default size
	//   - [AwaitableSlice.SetSize] was invoked with
	//	- — size argument between 1 and 10
	//	- — resulting allocation size is less then 4 KiB.
	//		size * sizeof T
	IsLowAlloc atomic.Bool
	// list provides: lock sliceList sliceList0
	list[T]
	// primary is primary input-queue value-slice
	//	- not a slice-away slice
	//	- only appended to when s.queue.sliceList is empty
	//	- may be empty while queue.sliceList is non-empty
	primary []T
	// cachedInput is a pre-allocated slice that me be used as
	// primary or value-slice in sliceList
	//	- saves on allocations behind inQ lock
	//	- copied from cachedOutput while holding both locks
	//	- used by Send SendClone methods
	//	- may be set by SendSlice or SendSlices
	cachedInput []T
}

// send enqueues value
func (i *inputQueue[T]) send(value T) {

	if i.IsLength.Load() {
		var length = i.Length.Add(1)
		i.MaxLength.Value(uint64(length))
	}

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
		var sliceList = i.makeSliceList(noMinSize, valueSlice)
		i.setList(sliceList)
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

// sendSlice enqueues by transferring ownership of
// values slice to the queue
func (i *inputQueue[T]) sendSlice(values []T) {

	if i.IsLength.Load() {
		var length = i.Length.Add(len(values))
		i.MaxLength.Value(uint64(length))
	}

	// create primary or replace empty primary
	//	- if primary length is zero, list is empty or unallocated
	if len(i.primary) == 0 {
		i.setAndDiscardPrimary(values)
		return // replaced primary queue return
	}
	// add to sliceList

	// create new sliceList
	if !i.HasList.Load() {
		var sliceList = i.makeSliceList(noMinSize, values)
		i.setList(sliceList)
		i.HasList.Store(true)
		return // created new sliceList return
	}

	// append to existing sliceList
	i.enqueueInList(values)
}

// SendClone enqueues a value-slice without transferring values slice ownership
// to the queue
func (i *inputQueue[T]) sendClone(values []T) {
	if i.IsLength.Load() {
		var length = i.Length.Add(len(values))
		i.MaxLength.Value(uint64(length))
	}

	var length, capacity = i.getPrimaryMetrics()

	// create primary
	if capacity == 0 {
		var valueSlice = i.createInputSlice(values...)
		i.primary = valueSlice
		return // created primary
	}
	// primary is allocated

	// append to primary if sliceList is empty
	var listLength, listCapacity = i.getListMetrics()
	if listLength == 0 {

		// copy to primary without allocation
		if length+len(values) <= capacity {
			i.enqueueInPrimary(values...)
			return // copy moved all to primary
		}

		// append up to max
		var primary = i.primary
		if i.appendToValueSlice(&primary, &values) {
			i.primary = primary
			if len(values) == 0 {
				return // all appended with realloc to q
			}
		}
	}
	// add values to sliceList

	// create sliceList
	if listCapacity == 0 {
		var valueSlice = i.createInputSlice(values...)
		var sliceList = i.makeSliceList(noMinSize, valueSlice)
		i.setList(sliceList)
		i.HasList.Store(true)
		return // created new sliceList return
	}

	// create new last-slice
	if listLength == 0 {
		var valueSlice = i.createInputSlice(values...)
		i.enqueueInList(valueSlice)
		return // created first sliceList element return
	}

	// append to last slice without allocation
	length, capacity = i.getLastSliceMetrics()
	if length+len(values) <= capacity {
		// copy without allocation
		i.enqueueInLastSlice(values...)
		return // copy moved all to last slice return
	}

	// append up to max with allocation
	var q = i.getLastSliceSlice()
	if i.appendToValueSlice(&q, &values) {
		i.swapLastSlice(q)
		if len(values) == 0 {
			return // all appended with realloc to last slice return
		}
	}

	// append new last slice
	var valueSlice = i.createInputSlice(values...)
	i.enqueueInList(valueSlice)
}

// sendSlices enqueues by transferring ownership of a list of slices to the queue
func (i *inputQueue[T]) sendSlices(valueSlices [][]T) {

	if i.IsLength.Load() {
		var length int
		for i := range valueSlices {
			length += len(valueSlices[i])
		}
		length = i.Length.Add(length)
		i.MaxLength.Value(uint64(length))
	}

	var startIndex int

	// if primary is empty, store first slice there
	if len(i.primary) == 0 {
		i.setAndDiscardPrimary(valueSlices[0])
		if len(valueSlices) == 1 {
			return // moved to primary queue return
		}
		startIndex = 1
	}

	// create sliceList from valueSlices
	if !i.HasList.Load() {
		i.setList(valueSlices)
		i.HasList.Store(true)
		if startIndex == 0 {
			return
		}
		// drop the slice now in primary
		i.dequeueFromList()
		return
	}

	// append value-slices to sliceList
	if startIndex == 0 {
		i.enqueueInList(valueSlices...)
		return
	}
	i.enqueueInList(valueSlices[startIndex:]...)
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

// setPrimary sets primary input queue
func (i *inputQueue[T]) setPrimary(primary []T) {
	i.primary = primary
}

// getPrimary gets primary input queue and its ownership
func (i *inputQueue[T]) getPrimary() (primary []T) {
	primary = i.primary
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
//   - used when transfering preallocated slices from outQ
func (i *inputQueue[T]) setCachedInput(slice ...[]T) {

	i.cachedInput = slice[0]
	var has = i.cachedInput != nil
	if has != i.HasInput.Load() {
		i.HasInput.Store(has)
	}
}

// setAndDiscardPrimary sets new primary and
// tries to retain primary as inputCached
//   - on SendSlice or Send Slices when primary is empty,
//     the primary slice may be saved to cachedInput
func (i *inputQueue[T]) setAndDiscardPrimary(slice []T) {
	if i.cachedInput == nil && cap(i.primary) > 0 && cap(i.primary) <= i.MaxRetainSize.Load() {
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
func (i *inputQueue[T]) makeValueSlice(values ...T) (newSlice []T) {

	// ensure size sizeMax are initialized
	var size = max(i.Size.Load(), len(values))

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
//   - minSize: minimum size or [noMinSize]
//   - slice: optional single slice to be first element
func (i *inputQueue[T]) makeSliceList(minSize int, slice ...[]T) (newSliceList [][]T) {

	var length int
	if len(slice) > 0 {
		length = 1
	}

	if i.IsLowAlloc.Load() {
		if minSize < lowAllocListSize {
			minSize = lowAllocListSize
		}
	} else if minSize < sliceListSize {
		minSize = sliceListSize
	}

	newSliceList = make([][]T, length, minSize)
	if len(slice) > 0 {
		newSliceList[0] = slice[0]
	}
	return
}

// appendToValueSlice appends values to valueSlice imited by max
//   - valueSlice was determine to not have enough capacity
func (i *inputQueue[T]) appendToValueSlice(valueSlice, values *[]T) (didChange bool) {
	var v = *valueSlice
	var vals = *values
	var nv0 = len(v)

	// if capacity too large, only copy
	if cap(v) >= maxAppendValueSliceCapacity {
		// n is number of elements available for copy
		var n = cap(v) - nv0
		if n == 0 {
			return // no additional elements
		}
		didChange = true
		n = min(n, len(vals))
		// extend v
		v = v[:nv0+n]
		copy(v[nv0:], vals[:n])
		*valueSlice = v
		if n == len(vals) {
			*values = nil
			return // entire values via copy
		}
		*values = vals[n:]
		return // some of values copied
	}
	// valueSlice can be extended
	didChange = true

	// n is initial allowed append length,
	// depends on current length
	var n = maxAppendValueSliceCapacity - len(v)
	// adjust for vals length
	var n0 = min(n, len(vals))
	v = append(v, vals[:n0]...)
	if len(vals) == n0 {
		*valueSlice = v
		*values = nil
		return // all appended return
	}
	vals = vals[n0:]
	// there are more in vals

	// copy any extra from excessive realloc
	n = cap(v) - len(v)
	if n == 0 {
		return
	}
	nv0 = len(v)
	n0 = min(n, len(vals))
	v = v[:nv0+n0]
	copy(v[nv0:], vals[:n0])
	*valueSlice = v
	if n0 == len(vals) {
		*values = nil
		return
	}
	*values = vals[n0:]

	return
}
