/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package ghandi interfaces Android devices
package threadprof

import (
	"sync/atomic"

	"github.com/haraldrudell/parl"
)

type Counter struct {
	value   uint64 // atomic
	running uint64 // atomic
	max     uint64 // atomic
	incRate uint64 // atomic
	decRate uint64 // atomic
}

func newCounter() (counter parl.Counter) {
	return &Counter{}
}

func (cn *Counter) Inc() (counter parl.Counter) {
	value := atomic.AddUint64(&cn.value, 1)
	atomic.AddUint64(&cn.running, 1)
	for {
		max := atomic.LoadUint64(&cn.max)
		if value <= max || // no update is required
			atomic.CompareAndSwapUint64(&cn.max, max, value) { // update was successful
			break
		}
	}
	return cn
}

func (cn *Counter) Dec() (cunter parl.Counter) {
	atomic.AddUint64(&cn.running, ^uint64(0))
	return cn
}

func (cn *Counter) CounterValue(reset bool) (values parl.CounterValues) {
	values = &CounterValue{
		value:   atomic.LoadUint64(&cn.value),
		running: atomic.LoadUint64(&cn.running),
		max:     atomic.LoadUint64(&cn.max),
		incRate: atomic.LoadUint64(&cn.incRate),
		decRate: atomic.LoadUint64(&cn.decRate),
	}
	if reset {
		atomic.StoreUint64(&cn.value, 0)
		atomic.StoreUint64(&cn.running, 0)
		atomic.StoreUint64(&cn.max, 0)
		atomic.StoreUint64(&cn.incRate, 0)
		atomic.StoreUint64(&cn.decRate, 0)
	}
	return
}
