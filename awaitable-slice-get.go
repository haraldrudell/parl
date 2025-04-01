/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "io"

// Get returns one value if the queue is not empty
//   - value: possible value
//   - hasValue true: value is valid
//   - hasValue false: the queue is empty, no value is provided
//   - Get may attain allocation-free receive or allocation-free operation
//   - — a slice is not returned
//   - — an internal slice may be reused reducing allocations
//   - thread-safe
func (s *AwaitableSlice[T]) Get() (value T, hasValue bool) {

	// fast check outside lock
	if !s.outQ.HasDataBits.hasData() {
		return
	}

	// checkedQueue is set to true if inQ was accessed
	//	- hasValue is true if a value was obtained
	//	- if a value was obtained without accessing inQ,
	//		the queue must be checked for now being empty
	var checkedQueue bool
	defer s.enterOutputCritical().postOutput(&hasValue, &checkedQueue)

	// check inside lock
	if !s.outQ.HasDataBits.hasData() {
		return
	}
	// there is at least one value in inQ or outQ

	return s.outQ.get(&checkedQueue)
}

// GetSlice returns a slice of values from the queue
//   - values non-nil: a non-empty slice at a time, not necessarily all data.
//     values is never non-nil and empty
//   - — Send-GetSlice: each GetSlice empties the queue
//   - — SendMany-GetSlice: each GetSlice receives one SendMany slice
//   - values nil: the queue is empty
//   - GetSlice may increase performance by slice-at-a-time operation, however,
//     slices need to be allocated:
//   - — Send-GetSlice requires internal slice allocation
//   - — SendMany-GetSlice requires sender to allocate slices
//   - — Send-Get1 may reduce allocations
//   - thread-safe
func (s *AwaitableSlice[T]) GetSlice() (values []T) {

	// fast check outside lock
	if !s.outQ.HasDataBits.hasData() {
		return // queue empty return
	}
	var hasValue, checkedQueue bool
	defer s.enterOutputCritical().postOutput(&hasValue, &checkedQueue)

	// check inside lock
	if !s.outQ.HasDataBits.hasData() {
		return // queue empty return
	}
	// there is at least one value in inQ or outQ
	hasValue = true

	return s.outQ.getSlice(&checkedQueue)
}

// GetSlices empties the queue at near zero allocations
//   - buffer missing: slices is allocated to exact size
//   - buffer: optional buffer to avoid allocation, typically length zero
//   - — capacity 10 or 100 can be used
//     Exact required size is not known prior to acquiring locks
//   - slices: list of non-empty value slices
//   - slices nil: the queue was empty
//   - thread-safe
func (s *AwaitableSlice[T]) GetSlices(buffer ...[][]T) (slices [][]T) {

	// fast check outside lock
	if !s.outQ.HasDataBits.hasData() {
		return // queue empty return
	}

	// GetSlices always sets hasData to false
	//	- disable postGet action
	var checkedQueue = true
	defer s.enterOutputCritical().postOutput(&checkedQueue, &checkedQueue)

	// check inside lock
	if !s.outQ.HasDataBits.hasData() {
		return // queue empty return
	}
	// there is at least one value in inQ or outQ

	var b [][]T
	if len(buffer) > 0 {
		b = buffer[0]
	}

	return s.outQ.getSlices(b)
}

// GetAll returns a single slice of all values in queue
//   - values nil: the queue is empty
//   - thread-safe
func (s *AwaitableSlice[T]) GetAll() (values []T) {

	// fast check outside lock
	if !s.outQ.HasDataBits.hasData() {
		return // queue empty return
	}

	// GetAll always sets hasData to false
	//	- disable postGet action
	var checkedQueue = true
	defer s.enterOutputCritical().postOutput(&checkedQueue, &checkedQueue)

	// check inside lock
	if !s.outQ.HasDataBits.hasData() {
		return // queue empty return
	}
	// there is at least one value in inQ or outQ

	return s.outQ.getAll()
}

// Read makes AwaitableSlice [io.Reader]
//   - p: buffer to read items to
//   - n: number of items read
//   - err: possible [io.EOF] once closed and read to empty
//   - — EOF may be returned along with read items
//   - —
//   - useful when a limited number of items is to be read, say 100 elements
//   - Read is non-blocking.
//     To block use [AwaitableSlice.DataWaitCh] or [AwaitableSlice.AwaitValue]
//   - non-blocking means that while the queue is empty, Read returns zero items
//   - may be less efficient that [AwaitableSlice.Get] or [AwaitableSlice.GetSlice]
//     by discarding internal slices
//   - Read combines data transfer with close which the slice otherwise treats separately
//   - if further data is written to slice after EOF using other than Write,
//     Read provides additional data after returning EOF.
//     This is because Send SendSlice SendSlices SendClone
//     continue to work after Close
//   - thread-safe
func (s *AwaitableSlice[T]) Read(p []T) (n int, err error) {

	// fast check outside lock
	if !s.outQ.HasDataBits.hasData() {
		if s.isCloseInvoked.Load() {
			err = io.EOF
		}
		return // slice empty return
	} else if len(p) == 0 {
		return // buffer zero-length return
	}
	// slice was observed to have data and buffer is of non-zero length

	// Read will update hasData so disable postGet action
	var checkedQueue = true
	defer s.enterOutputCritical().postOutput(&checkedQueue, &checkedQueue)

	// check inside lock
	if !s.outQ.HasDataBits.hasData() {
		if s.isCloseInvoked.Load() {
			err = io.EOF
		}
		return // slice empty return
	}
	// there is data in inQ or outQ:
	// at least one value will be returned

	n, err = s.outQ.read(p)

	if !s.outQ.HasDataBits.hasData() && s.isCloseInvoked.Load() {
		err = io.EOF
	}

	return
}

// AwaitValue awaits value or close, blocking until either event
//   - hasValue true: value is valid, possibly the zero-value like
//     a nil interface value
//   - hasValue: false: the stream is closed
//   - stream: an awaitable possibly closable source type like [Source1]
//   - — stream’s DataWaitCh Get and if present EmptyCh methods are used
//   - — stream cannot be eg. [AtomicError] because it is not awaitable
//   - AwaitValue wraps a 10-line read operation as a two-value expression
//   - thread-safe
func (s *AwaitableSlice[T]) AwaitValue() (value T, hasValue bool) {

	// loop until value or closed
	var endCh AwaitableCh
	var dataWait *CyclicAwaitable
	for {

		// competing with other threads for values
		//	- may receive nothing
		if value, hasValue = s.Get(); hasValue {
			return // value read return: hasValue true, value valid
		}

		if endCh == nil {
			endCh, dataWait = s.getAwait()
		}
		select {
		case <-endCh:
			return // closable is closed return: hasValue false
		case <-dataWait.Ch():
		}
	}
}

// Seq allows for AwaitableSlice to be used in a for range clause.
// Seq blocks if queue is empty and not closed
//   - each value is provided to yield
//   - iterates until yield retuns false or
//   - the slice was empty and in drain-close states
//   - thread-safe
//
// Usage:
//
//	for value := range awaitableSlice.Seq {
//	  value…
//	}
//	// the AwaitableSlice was empty and in drain-closed state
func (s *AwaitableSlice[T]) Seq(yield func(value T) (keepGoing bool)) {

	// loop until value or closed
	var endCh AwaitableCh
	var dataWait *CyclicAwaitable
	var value T
	var hasValue bool
	for {

		// competing with other threads for values
		//	- may receive nothing
		if value, hasValue = s.Get(); hasValue {
			if !yield(value) {
				return // consumer canceled iteration return
			}
			continue
		}

		if endCh == nil {
			endCh, dataWait = s.getAwait()
		}
		select {
		case <-endCh:
			return // closable is closed return: hasValue false
		case <-dataWait.Ch():
		}
	}
}
