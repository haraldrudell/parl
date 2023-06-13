/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"
	"time"
)

// Debouncer debounces event stream values.
// T values are received from the in channel.
// Once d time has elapsed with no further incoming Ts,
// a slice of read Ts are provided to the send function.
//   - the debouncer may be held up indefinitely for an uninterrupted stream of Ts
//   - two threads are launched per debouncer
//   - errFn receives any panics in the threads
//   - sender and errFn functions must be thread-safe.
//   - Debouncer is shutdown gracefully by input channel close or
//     immediately using Shutdown method
type Debouncer[T any] struct {
	duration                            time.Duration
	inputCh                             <-chan T
	buffer                              NBChan[T] // non-blocking unbound buffer
	timer                               *time.Timer
	sender                              func([]T)
	errFn                               func(err error)
	inputEndCh, outputEndCh, shutdownCh chan struct{}
	isShutdown                          atomic.Bool
}

// NewDebouncer returns a channel debouncer
func NewDebouncer[T any](
	duration time.Duration,
	inputCh <-chan T,
	sender func([]T),
	errFn func(err error)) (debouncer *Debouncer[T]) {
	db := Debouncer[T]{
		duration:    duration,
		inputCh:     inputCh,
		sender:      sender,
		errFn:       errFn,
		timer:       time.NewTimer(time.Second),
		inputEndCh:  make(chan struct{}),
		outputEndCh: make(chan struct{}),
		shutdownCh:  make(chan struct{}),
	}
	db.timer.Stop()
	if len(db.timer.C) > 0 {
		<-db.timer.C
	}
	return &db
}

// Go launches the debouncer thread
func (d *Debouncer[T]) Go() (debouncer *Debouncer[T]) {
	debouncer = d
	go d.inputThread()
	go d.outputThread()
	return
}

func (d *Debouncer[T]) Shutdown() {
	if d.isShutdown.CompareAndSwap(false, true) {
		close(d.shutdownCh)
	}
	d.Wait()
}

// Wait blocks until the debouncer exits
//   - the debouncer exits from in channel close or context cancel
func (d *Debouncer[T]) Wait() {
	<-d.inputEndCh
	<-d.outputEndCh
}

// debouncerThread debounces the in channel until it closes or context cancel
func (d *Debouncer[T]) inputThread() {
	defer close(d.inputEndCh)
	defer Recover(Annotation(), nil, d.errFn)

	// read input channel save in buffer and reset timer
	var noShutdown = true
	for {
		var value T
		var ok bool
		select {
		case value, ok = <-d.inputCh:
		case _, noShutdown = <-d.shutdownCh:
		}
		if !ok || !noShutdown {
			d.buffer.Close()
			return // input channel closed return
		}
		d.buffer.Send(value)
		// Stop prevents the Timer from firing
		d.timer.Stop()
		// drain the channel without blocking
		if len(d.timer.C) > 0 {
			select {
			case <-d.timer.C:
			default:
			}
		}
		// Reset should be invoked only on:
		//	- stopped or expired timers
		//	- with drained channels
		d.timer.Reset(d.duration)
	}
}

// debouncerThread debounces the in channel until it closes or context cancel
func (d *Debouncer[T]) outputThread() {
	defer close(d.outputEndCh)
	defer Recover(Annotation(), nil, d.errFn)

	// wait for timer to elapse or input thread to exit
	var inputThreadOK = true
	var noShutdown = true
	for {
		select {
		case <-d.timer.C: // wait for debounce time
		case _, inputThreadOK = <-d.inputEndCh: // wait for input thread to exit
		case _, noShutdown = <-d.shutdownCh:
		}
		if !noShutdown {
			return // shutdown received
		}
		if values := d.buffer.Get(); len(values) > 0 {
			d.sender(values)
		}
		if !inputThreadOK {
			return // input thread ended and buffer is empty return
		}
	}
}
