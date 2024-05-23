/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"github.com/haraldrudell/parl/internal/cyclebreaker"
)

// RangeCh is a range-able channel based on a thread-safe slice
//   - the slice is provided to the New function
//   - each RangeCh has an internal thread that runs until:
//   - — the slice is out of items
//   - — Close is invoked
//   - RangeCh.Ch() must be either read until close or RangeCh.Close must be invoked
//   - Close discards items
//   - —
//   - it is unclear whether pairing an updatable slice with an unbuffered channel is useful
//   - — unlike a buffered channel, RangeCh is unbound
//   - — RangeCh costs 1 thread maintaining the unbuffered channel
//   - — channels are somewhat inefficient by processing only one element at a time
//   - — controlling allocations by processing single elements, slices of elements or
//     slices of slices of elements
//   - — RangeCh sends any available elements then closes the channel and exits
type RangeCh[T any] struct {
	// the range-able channel sending data
	//	- closes on end-of-data
	//	- unbuffered
	//	- send and close by rangeChWriteThread
	ch chan T
	// isClosed indicates that [RangeCh.Close] has been invoked
	//	- the channel ch is about to close
	//	- isClosed.Close is idempotent
	//	- synchronization mechanic to exit rangeChWriteThread
	//	- observable
	isClosed cyclebreaker.Awaitable
	// isExit is synchronization mechanic that rangeChWriteThread has exit
	//	- on isExit closing, ch is already closed
	isExit cyclebreaker.AwaitableCh
}

// NewRangeCh returns a range-able channel based on a [ThreadSafeSlice]
//   - the Ch channel must be either read until it closes or Close must be invoked
//   - Ch is a channel that can be used in a Go for-range clause
//   - —
//   - for can range over: array slice string map or channel
//   - the only dynamic range source is channel which costs a thread sending on and closing the channel
func NewRangeCh[T any](tss *ThreadSafeSlice[T]) (rangeChan *RangeCh[T]) {
	var isExit = make(chan struct{})
	r := RangeCh[T]{
		ch:     make(chan T),
		isExit: isExit,
	}

	// launch sending thread
	go rangeChWriteThread(tss, r.ch, r.isClosed.Ch(), isExit)

	return &r
}

// Ch returns the range-able channel sending values, thread-safe
func (r *RangeCh[T]) Ch() (ch <-chan T) { return r.ch }

// Close closes the ch channel, thread-safe, idempotent
//   - Close may discard a pending item and causes the thread to stop reading from the slice
//   - Close does not return until Ch is closed and the thread have exited
func (r *RangeCh[T]) Close() {

	// cause rangeChWriteThread to exit
	r.isClosed.Close()

	// wait for rangeChWriteThread to exit
	<-r.isExit
}

// State returns [RangeCh] state
// - isCloseInvoked is true if [RangeCh.Close] has been invoked, but Close may not have completed
// - isChClosed means [RangeCh.Ch] has closed and resources are released
// - exitCh allows to wait for Close complete
func (r *RangeCh[T]) State() (isCloseInvoked, isChClosed bool, exitCh <-chan struct{}) {
	isCloseInvoked = r.isClosed.IsClosed()
	exitCh = r.isExit
	select {
	case <-exitCh:
		isChClosed = true
	default:
	}
	return
}

// rangeChWriteThread writes elements to r.ch until end or Close
//   - typically rangeChWriteThread will block in channel send
//   - infallible: runtime errors printed to standard error
func rangeChWriteThread[T any](
	tss *ThreadSafeSlice[T],
	ch chan<- T,
	isClosedCh cyclebreaker.AwaitableCh,
	isExit chan struct{},
) {
	var err error
	defer cyclebreaker.Recover(func() cyclebreaker.DA { return cyclebreaker.A() }, &err, cyclebreaker.Infallible)
	defer cyclebreaker.Closer(isExit, &err)
	defer cyclebreaker.CloserSend(ch, &err)

	var index int
	for {

		// get element to send
		//	- non-blocking
		//	- hasValue false when out of values
		var element, hasValue = tss.Get(index)
		if !hasValue {
			return // end of items return
		}
		index++

		// write thread normally blocks here
		select {
		case ch <- element:
		case <-isClosedCh:
			return // [RangeCh.Close] invoked
		}
	}
}
