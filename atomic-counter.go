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

func (max *AtomicCounter) Dec() (value uint64) {
	value = atomic.AddUint64((*uint64)(max), math.MaxUint64)
	return
}

func (max *AtomicCounter) Add(value uint64) (newValue uint64) {
	newValue = atomic.AddUint64((*uint64)(max), value)
	return
}

func (max *AtomicCounter) Value() (value uint64) {
	value = atomic.LoadUint64((*uint64)(max))
	return
}
