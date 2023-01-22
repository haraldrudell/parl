/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
	"sync/atomic"

	"golang.org/x/exp/constraints"
)

const (
	stateUninitialized = 0
	stateHasValue      = 1
)

type AtomicMin[T constraints.Integer] struct {
	state uint32
	once  sync.Once
	value uint64
}

func (min *AtomicMin[T]) Value(value T) (isNewMin bool) {

	var valueU64 uint64 = uint64(value)

	// ensure initialized
	if atomic.LoadUint32(&min.state) == stateUninitialized {
		min.once.Do(func() {
			atomic.StoreUint64(&min.value, valueU64)
			atomic.StoreUint32(&min.state, stateHasValue)
			isNewMin = true
		})
		if isNewMin {
			return // value-initializing invocation always has min value
		}
	}

	// aggregate minimum
	var current uint64 = atomic.LoadUint64(&min.value)
	if isNewMin = valueU64 < current; !isNewMin {
		return // too large value, nothing to do return
	}

	// ensure write of new min value
	for {

		// try to write
		if atomic.CompareAndSwapUint64(&min.value, current, valueU64) {
			return // min.value updated return
		}

		// load new copy of value
		current = atomic.LoadUint64(&min.value)
		if current <= valueU64 {
			return // min.value now ok return
		}
	}
}

func (min *AtomicMin[T]) Min() (value T, hasValue bool) {
	if hasValue = atomic.LoadUint32(&min.state) == stateHasValue; hasValue {
		value = T(atomic.LoadUint64(&min.value))
	}
	return
}
