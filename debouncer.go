/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"sync"
	"time"
)

// Debouncer debounces event stream values.
// T values are received from the in channel.
// once d time has elapsed with no further incoming Ts,
// a slice of read Ts are provided to the send function.
//   - the debouncer may be held up indefinitely for an uninterrupted stream of Ts
//   - one thread is launched per debouncer
//   - errFn receives any panics in the thread
//   - sender and errFb functions must be thread-safe.
//   - Debouncer is shutdown by in channel close or via context.
type Debouncer[T any] struct {
	d      time.Duration
	in     <-chan T
	sender func([]T)
	errFn  func(err error)
	wg     sync.WaitGroup
	ctx    context.Context
}

// NewDebouncer returns a channel debouncer
func NewDebouncer[T any](d time.Duration, in <-chan T, sender func([]T),
	errFn func(err error), ctx context.Context) (db *Debouncer[T]) {
	return &Debouncer[T]{
		d:      d,
		in:     in,
		sender: sender,
		errFn:  errFn,
		ctx:    ctx,
	}
}

// Go launches the debouncer thread
func (d *Debouncer[T]) Go() {
	d.wg.Add(1)
	go d.debouncerThread()
}

// Wait blocks until the debouncer exits
//   - the debouncer exits from in channel close or context cancel
func (db *Debouncer[T]) Wait() {
	db.wg.Wait()
}

// debouncerThread debounces the in channel until it closes or context cancel
func (d *Debouncer[T]) debouncerThread() {
	defer d.wg.Done()
	Recover(Annotation(), nil, d.errFn)

	timer := time.NewTimer(time.Second)
	timer.Stop()
	defer timer.Stop()

	done := d.ctx.Done()
	var values []T
	for {
		select {
		case value, ok := <-d.in:
			if !ok {
				if len(values) > 0 {
					d.sender(values)
				}
				return // in closed return
			}
			values = append(values, value)
			timer.Reset(d.d)
		case <-done:
			return // ctx shutdown return
		case <-timer.C:
			d.sender(values)
			values = nil
		}
	}
}
