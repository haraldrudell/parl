/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"math"
	"sync/atomic"
)

type AtomicCounter uint64

func (max *AtomicCounter) Inc() (value uint64) {
	value = atomic.AddUint64((*uint64)(max), 1)
	return
}

func (max *AtomicCounter) Inc2() (value uint64, didInc bool) {
	for {
		var beforeValue = atomic.LoadUint64((*uint64)(max))
		if beforeValue == math.MaxUint64 {
			return
		} else if didInc = atomic.CompareAndSwapUint64((*uint64)(max), beforeValue, beforeValue+1); didInc {
			return
		}
	}
}

func (max *AtomicCounter) Dec() (value uint64) {
	value = atomic.AddUint64((*uint64)(max), math.MaxUint64)
	return
}

func (max *AtomicCounter) Dec2() (value uint64, didDec bool) {
	for {
		var beforeValue = atomic.LoadUint64((*uint64)(max))
		if beforeValue == 0 {
			return
		} else if didDec = atomic.CompareAndSwapUint64((*uint64)(max), beforeValue, beforeValue-1); didDec {
			return
		}
	}
}

func (max *AtomicCounter) Add(value uint64) (newValue uint64) {
	newValue = atomic.AddUint64((*uint64)(max), value)
	return
}

func (max *AtomicCounter) Set(value uint64) (oldValue uint64) {
	oldValue = atomic.SwapUint64((*uint64)(max), value)
	return
}

func (max *AtomicCounter) Value() (value uint64) {
	value = atomic.LoadUint64((*uint64)(max))
	return
}
