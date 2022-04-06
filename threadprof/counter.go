/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

// Package ghandi interfaces Android devices
package threadprof

import "sync/atomic"

type CounterOn struct {
	value   uint64 // atomic
	ops     uint64 // atomic
	max     uint64 // atomic
	incRate uint64 // atomic
	decRate uint64 // atomic
}

func newCounter() (counter Counter) {
	return &CounterOn{}
}

func (cn *CounterOn) Inc() (value uint64) {
	value = atomic.AddUint64(&cn.value, 1)
	atomic.AddUint64(&cn.ops, 1)
	for {
		max := atomic.LoadUint64(&cn.max)
		if value <= max || // no update is required
			atomic.CompareAndSwapUint64(&cn.max, max, value) { // update was successful
			break
		}
	}
	return
}

func (cn *CounterOn) Dec() (value uint64) {
	value = atomic.AddUint64(&cn.value, ^uint64(0))
	atomic.AddUint64(&cn.ops, 1)
	return
}

func (cn *CounterOn) CounterValue(reset bool) (values Values) {
	values = &CounterValue{
		value:   atomic.LoadUint64(&cn.value),
		ops:     atomic.LoadUint64(&cn.ops),
		max:     atomic.LoadUint64(&cn.max),
		incRate: atomic.LoadUint64(&cn.incRate),
		decRate: atomic.LoadUint64(&cn.decRate),
	}
	if reset {
		atomic.StoreUint64(&cn.value, 0)
		atomic.StoreUint64(&cn.ops, 0)
		atomic.StoreUint64(&cn.max, 0)
	}
	return
}
