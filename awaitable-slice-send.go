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

	s.outQ.InQ.send(value)
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

	s.outQ.InQ.sendSlice(values)
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

	s.outQ.InQ.sendClone(values)
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

	s.outQ.InQ.sendSlices(valueSlices)
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
