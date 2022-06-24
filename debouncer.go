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

type Debouncer[T any] struct {
	d      time.Duration
	in     <-chan T
	sender func([]T)
	errFn  func(err error)
	wg     sync.WaitGroup
	ctx    context.Context
}

// Debouncer debounces event stream values.
// sender and errFb functions must be thread-safe.
// Debouncer is shutdown by in channel close or via context.
func NewDebouncer[T any](d time.Duration, in <-chan T,
	sender func([]T),
	errFn func(err error), ctx context.Context) (db *Debouncer[T]) {
	db0 := Debouncer[T]{d: d, in: in, sender: sender, ctx: ctx}
	db0.wg.Add(1)
	go db0.debouncerThread()
	return &db0
}

func (db *Debouncer[T]) Wait() {
	db.wg.Wait()
}

func (db *Debouncer[T]) debouncerThread() {
	defer db.wg.Done()
	Recover(Annotation(), nil, db.errFn)

	timer := time.NewTimer(time.Second)
	timer.Stop()
	defer timer.Stop()

	done := db.ctx.Done()
	var values []T
	for {
		select {
		case value, ok := <-db.in:
			if !ok {
				if len(values) > 0 {
					db.sender(values)
				}
				return // in closed return
			}
			values = append(values, value)
			timer.Reset(db.d)
		case <-done:
			return // ctx shutdown return
		case <-timer.C:
			db.sender(values)
			values = nil
		}
	}
}
