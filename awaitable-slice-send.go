/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "os"

// Send enqueues a single value
//   - panic-free error-free thread-safe
func (s *AwaitableSlice[T]) Send(value T) {
	defer s.enterInputCritical().postInput()
	s.inQ.send(value)
}

// SendSlice enqueues by transferring ownership of values slice to the queue
//   - SendSlice may reduce allocations and increase performance by handling multiple values
//   - panic-free error-free thread-safe
func (s *AwaitableSlice[T]) SendSlice(values []T) {

	// ignore empty slice
	if len(values) == 0 {
		return
	}
	defer s.enterInputCritical().postInput()

	// create primary or replace empty primary
	//	- if primary length is zero, list is empty or unallocated
	if s.inQ.isEmptyPrimary() {
		s.inQ.setAndDiscardPrimary(values)
		return // replaced primary queue return
	}
	// add to sliceList

	// create new sliceList
	if !s.inQ.HasList.Load() {
		var sliceList = s.inQ.makeSliceList(values)
		s.inQ.swapList(sliceList)
		s.inQ.HasList.Store(true)
		return // created new sliceList return
	}

	// append to existing sliceList
	s.inQ.enqueueInList(values)
}

// SendClone enqueues a value-slice without transferring values slice ownership
// to the queue
//   - panic-free error-free thread-safe
func (s *AwaitableSlice[T]) SendClone(values []T) {

	// ignore empty slice
	if len(values) == 0 {
		return
	}
	defer s.enterInputCritical().postInput()

	var length, capacity = s.inQ.getPrimaryMetrics()

	// create primary
	if capacity == 0 {
		var valueSlice = s.inQ.createInputSlice(values...)
		s.inQ.swapPrimary(valueSlice)
		return // created primary
	}
	// primary is allocated

	// append to primary if sliceList is empty
	var listLength, listCapacity = s.inQ.getListMetrics()
	if listLength == 0 {

		// copy to primary without allocation
		if length+len(values) <= capacity {
			s.inQ.enqueueInPrimary(values...)
			return // copy moved all to primary
		}

		// append up to max
		var primary = s.inQ.getPrimarySlice()
		if s.appendToValueSlice(&primary, &values) {
			s.inQ.swapPrimary(primary)
			if len(values) == 0 {
				return // all appended with realloc to q
			}
		}
	}
	// add values to sliceList

	// create sliceList
	if listCapacity == 0 {
		var valueSlice = s.inQ.createInputSlice(values...)
		var sliceList = s.inQ.makeSliceList(valueSlice)
		s.inQ.swapList(sliceList)
		s.inQ.HasList.Store(true)
		return // created new sliceList return
	}

	// create new last-slice
	if listLength == 0 {
		var valueSlice = s.inQ.createInputSlice(values...)
		s.inQ.enqueueInList(valueSlice)
		return // created first sliceList element return
	}

	// append to last slice without allocation
	length, capacity = s.inQ.getLastSliceMetrics()
	if length+len(values) <= capacity {
		// copy without allocation
		s.inQ.enqueueInLastSlice(values...)
		return // copy moved all to last slice return
	}

	// append up to max with allocation
	var q = s.inQ.getLastSliceSlice()
	if s.appendToValueSlice(&q, &values) {
		s.inQ.swapLastSlice(q)
		if len(values) == 0 {
			return // all appended with realloc to last slice return
		}
	}

	// append new last slice
	var valueSlice = s.inQ.createInputSlice(values...)
	s.inQ.enqueueInList(valueSlice)
}

// SendSlices enqueues by transferring ownership of a list of slices to the queue
//   - SendSlice may reduce allocations and increase performance by handling multiple values
//   - panic-free error-free thread-safe
func (s *AwaitableSlice[T]) SendSlices(valueSlices [][]T) {

	// noop check
	if len(valueSlices) == 0 {
		return // empty valueSlices return
	}

	// filter valueSlices
	//	- empty slice entries are not allowed
	var put int
	for get := range len(valueSlices) {
		if len(valueSlices[get]) == 0 {
			continue // skip empty slices
		}
		if put < get {
			valueSlices[put] = valueSlices[get]
		}
		put++
	}
	if put == 0 {
		return // no non-empty slices return
	} else if put < len(valueSlices) {
		valueSlices = valueSlices[:put]
	}
	// valueSlices is non-empty
	defer s.enterInputCritical().postInput()

	var startIndex int

	// if primary is empty, store first slice there
	if s.inQ.isEmptyPrimary() {
		s.inQ.setAndDiscardPrimary(valueSlices[0])
		if len(valueSlices) == 1 {
			return // moved to primary queue return
		}
		startIndex = 1
	}

	// create sliceList from valueSlices
	if !s.inQ.HasList.Load() {
		s.inQ.swapList(valueSlices)
		s.inQ.HasList.Store(true)
		if startIndex == 0 {
			return
		}
		// drop the slice now in primary
		s.inQ.dequeueFromList()
		return
	}

	// append value-slices to sliceList
	if startIndex == 0 {
		s.inQ.enqueueInList(valueSlices...)
		return
	}
	s.inQ.enqueueInList(valueSlices[startIndex:]...)
}

// Write makes queue [io.Writer]
//   - err: only [os.ErrClosed] is [AwaitableSlice.Close] was invoked
//   - otherwise: n = len(p)
func (s *AwaitableSlice[T]) Write(p []T) (n int, err error) {

	// Write after Close is error
	if s.isCloseInvoked.Load() {
		err = os.ErrClosed
		return
	}

	// write empty slice is noop
	if len(p) == 0 {
		return
	}

	// write never fails
	n = len(p)
	s.SendClone(p)

	return
}
