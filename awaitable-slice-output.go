/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/pslices/pslib"

// outputQueue contains dequeueing lock and structures
type outputQueue[T any] struct {
	// queue provides: lock sliceList sliceList0
	list[T]
	// head is a slice being sliced away from
	//	- head is the result of a slice-expression on head0.
	//		head and head0 share underlying array
	//	- head is slice-away slice
	head []T
	// head0 is full-length slice returned by make
	//	- the purpose of head0 is to return head to
	//		the beginning of the underlying array
	head0 []T
	// a pre-allocated value-slice transfered to outputQueue Lock
	//	- allocated outside of outputQueue lock
	cachedOutput []T
	ZeroOut      Atomic32[pslib.ZeroOut]
}

// isEmptyOutput returns true if outputQueue is empty
func (o *outputQueue[T]) isEmptyOutput() (isEmpty bool) {
	return len(o.head) == 0 && len(o.sliceList) == 0
}

// isEmpty returns true if primary output queue is empty
func (o *outputQueue[T]) isEmptyHead() (isEmpty bool) {
	return len(o.head) == 0
}

// getCount returns the total number of elements in primary
// output queue
func (o *outputQueue[T]) getHeadMetrics() (length, capacity int) {
	length = len(o.head)
	capacity = cap(o.head)
	return
}

// getOutputSlice gets the output slice and its ownership
func (o *outputQueue[T]) swapHead(newHead ...[]T) (head, head0 []T) {

	// set case
	if len(newHead) > 0 {
		var h = newHead[0]
		o.head = h
		if c := cap(h); len(h) < c {
			h = h[:c]
		}
		o.head0 = h
		return
	}

	// get case
	head = o.head
	head0 = o.head0
	o.head = nil
	o.head0 = nil
	return
}

// dequeueFromHead retrieves value from output using slice-away
//   - output must have been verified to have non-zero length
func (o *outputQueue[T]) dequeueFromHead() (value T) {

	value = o.head[0]
	if o.ZeroOut.Load() != pslib.NoZeroOut {
		clear(o.head[0:1])
	}
	if len(o.head) == 1 {
		o.head = o.head0[:0]
	} else {
		o.head = o.head[1:]
	}
	return
}

// dequeueHead moves all elements from primary output
// to dest
//   - dest: destination slice
//   - — dest must be of sufficient length
//   - n: number of elements copies
func (o *outputQueue[T]) dequeueHead(dest []T) (n int) {
	n = copy(dest, o.head)
	if o.ZeroOut.Load() != pslib.NoZeroOut {
		clear(o.head)
	}
	o.head = o.head0[:0]
	return
}

// dequeueNFromOutput copies to dest from all slices
//   - useful to [io.Read]-type funcctions
func (o *outputQueue[T]) dequeueNFromOutput(dest *[]T, np *int) (isDone bool) {

	// o.output
	if len(o.head) > 0 {
		// copy with zero-out
		var zeroOut = o.ZeroOut.Load()
		isDone = CopyToSlice(&o.head, dest, np, zeroOut)
		if len(o.head) == 0 {
			o.head = o.head0[:0]
		}
		if isDone {
			return
		}
	}

	// sliceList
	return o.dequeueNFromList(dest, np, o.ZeroOut.Load())
}

// hasCachedOutput returns true if a slice has been cached
func (o *outputQueue[T]) hasCachedOutput() (has bool) {
	return o.cachedOutput != nil
}

// swapCachedOutput sets or gets cachedInput
func (o *outputQueue[T]) swapCachedOutput(slice ...[]T) (cached []T) {

	// set case
	if len(slice) > 0 {
		o.cachedOutput = slice[0]
		return
	}

	// get case
	cached = o.cachedOutput
	o.cachedOutput = nil

	return
}

// trySaveToCachedOutput saves the head slice as
// cahchedOutput if it is suitable
//   - maxCap: maximum capacity that is retained
func (o *outputQueue[T]) trySaveToCachedOutput(maxCap int) {
	if o.cachedOutput != nil {
		return // cached output is already present
	} else if c := cap(o.head0); c == 0 || c > maxCap {
		return // unsuitable slice
	}
	o.cachedOutput = o.head0
}
