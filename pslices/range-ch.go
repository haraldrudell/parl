/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
)

// RangeCh is a range-able channel based on a thread-safe slice
//   - RangeCh.Ch() must be either read until close or RangeCh.Close must be invoked
type RangeCh[T any] struct {
	ch <-chan T
	// r must be pointer since thread is launched in New function
	//	- otherwise, memory leak may result
	r *rangeCh[T]
}

type rangeCh[T any] struct {
	ch              chan T
	threadSafeSlice *ThreadSafeSlice[T]

	// thread-safe mutable fields
	isClosed  atomic.Bool
	closeOnce sync.Once
	wg        sync.WaitGroup

	// index belongs to writeThread
	index int
	// writeThread closes exitCh on exit
	exitCh chan struct{}
}

// NewRangeCh returns a range-able channel based on a ThreadSafeSlice
//   - the channel must be either read until it closes or invoke Close
func NewRangeCh[T any](tss *ThreadSafeSlice[T]) (rangeChan *RangeCh[T]) {

	// create local pointed-to structure
	var r = rangeCh[T]{
		ch:              make(chan T),
		threadSafeSlice: tss,
		index:           -1,
		exitCh:          make(chan struct{}),
	}
	r.wg.Add(1)
	go r.writeThread()

	return &RangeCh[T]{ch: r.ch, r: &r}
}

// Ch returns the channel, thread-safe
func (r *RangeCh[T]) Ch() (ch <-chan T) {
	return r.ch
}

// Close closes the channel, thread-safe, idempotent
func (r *RangeCh[T]) Close() {
	r.r.close()
}

// writeThread writes elements to r.ch until end or Close
func (r *rangeCh[T]) writeThread() {
	defer r.wg.Done()
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

func (r *rangeCh[T]) close() {
	r.closeOnce.Do(r.doClose)
}

func (r *rangeCh[T]) doClose() {

	// signal to writeThread to exit
	r.isClosed.Store(true)

	// read to ensure writeThread is unblocked
	select {
	case <-r.ch: // this unblocks writeThread so it finds r.isClosed to be true
	case <-r.exitCh: // this happens on writeThread exit
	}

	// wait for writeThread to exit
	r.wg.Wait()
}
