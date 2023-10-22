/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"sync/atomic"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
)

// RangeCh is a range-able channel based on a thread-safe slice
//   - the slice is provided to the New function
//   - each RangeCh has an internal thread that runs until:
//   - — the slice is out of items
//   - — Close is invoked
//   - RangeCh.Ch() must be either read until close or RangeCh.Close must be invoked
type RangeCh[T any] struct {
	// ref must be pointer since a thread is launched in New function
	//	- otherwise, memory leak may result
	ref *threadData[T]
}

// NewRangeCh returns a range-able channel based on a ThreadSafeSlice
//   - the Ch channel must be either read until it closes or Close must be invoked
//   - Ch is a channel that can be used in a for range clause
//   - for can range over: array slice string map or channel
//   - the only dynamic range source is channel which costs a thread writing to and closing the channel
func NewRangeCh[T any](tss *ThreadSafeSlice[T]) (rangeChan *RangeCh[T]) {

	// create local pointed-to structure
	//	- the pointer allows for launching a thread inside a new-function
	var t = threadData[T]{
		ch:              make(chan T),
		threadSafeSlice: tss,
		index:           -1,
		exitCh:          make(chan struct{}),
	}

	// launch sending thread
	go t.writeThread()

	return &RangeCh[T]{ref: &t}
}

// ref is an internal structure supporting writeThread
type threadData[T any] struct {
	// the range-able channel sending data
	//	- the thread sends on ch, so it must be the thread closing ch
	ch chan T
	// threadSafeSlice holds the data being sent
	threadSafeSlice *ThreadSafeSlice[T]

	// thread-safe mutable fields

	// isClosed indicates that Close has been invoked
	//	- the channel ch is about to close
	//	- signals to the thread to stop reading from the slice
	isClosed atomic.Bool

	// writeThread fields

	index int // next index in thread-safe slice
	// writeThread closes exitCh on exit
	//	- when thread closes exitCh, ch has already been closed
	exitCh chan struct{}
}

// Ch returns the range-able channel sending values, thread-safe
func (r *RangeCh[T]) Ch() (ch <-chan T) {
	return r.ref.ch
}

// Close closes the ch channel, thread-safe, idempotent
//   - Close may discard a pending item and causes the thread to stop reading from the slice
//   - Close does not return until Ch is closed and the thread have exited
func (r *RangeCh[T]) Close() {
	defer func() { <-r.ref.exitCh }() // wait for thread to close ch, then close exitCh, then exit

	// try to win the first-to-close race
	//	- r.isClosed true signals to writeThread to exit
	if !r.ref.isClosed.CompareAndSwap(false, true) {
		return // loser waits for winner
	}

	// read to ensure writeThread is unblocked
	select {
	case <-r.ref.ch: // this unblocks writeThread so it finds r.isClosed to be true
	default:
	}
}

// - isCloseInvoked is true if Close has been invoked, may not have returned or closed ch yet
// - isChClosed means Ch has closed
// - exitCh is a channel that closes, ie. it allows waiting for Ch to close
func (r *RangeCh[T]) State() (isCloseInvoked, isChClosed bool, exitCh <-chan struct{}) {
	isCloseInvoked = r.ref.isClosed.Load()
	exitCh = r.ref.exitCh
	select {
	case <-exitCh:
		isChClosed = true
	default:
	}
	return
}

// writeThread writes elements to r.ch until end or Close
//   - typically writeThread will block in channel send
func (r *threadData[T]) writeThread() {
	var err error
	defer cyclebreaker.Recover(cyclebreaker.Annotation(), &err, cyclebreaker.Infallible)
	defer cyclebreaker.Closer(r.exitCh, &err)
	defer cyclebreaker.Closer(r.ch, &err)

	var element T
	var hasValue bool
	for {

		// get element to send
		r.index++
		if element, hasValue = r.threadSafeSlice.Get(r.index); !hasValue {
			return // end of items return
		}

		// write thread normally blocks here
		r.ch <- element

		// check if Close was invoked while blocked
		if r.isClosed.Load() {
			return // closed by r.doClose return
		}
	}
}
