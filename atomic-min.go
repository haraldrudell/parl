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

// AtomicMin is a thread-safe container for a minimum value of any integer type
//   - hasValue indicator
//   - generic for any underlying Integer type
//   - if type is signed, min may be negative
//   - lock for first Value invocation
//   - initialization-free
type AtomicMin[T constraints.Integer] struct {
	isInitialized atomic.Bool   // whether a value is present
	value         atomic.Uint64 // current min value as uint64
	initLock      sync.Mutex    // thread selector and wait for wtriting initial value
}

// Value notes a new min-candidate
//   - if not a new minima, state is not changed
//   - Thread-safe
func (a *AtomicMin[T]) Value(value T) (isNewMin bool) {

	// value-valueU64 is candidate min-value
	var valueU64 uint64 = uint64(value)

	// ensure initialized
	if !a.isInitialized.Load() {
		if isNewMin = a.init(valueU64); isNewMin {
			return // this thread set initial value return
		}
	}

	// aggregate minimum
	var current = a.value.Load()
	var currentT = T(current)
	// make comparison in T domain
	if isNewMin = value < currentT; !isNewMin {
		return // too large value, nothing to do return
	}

	// ensure write of new min value
	for {

		// try to write
		if a.value.CompareAndSwap(current, valueU64) {
			return // min-value updated return
		}

		// load new copy of value
		current = a.value.Load()
		currentT = T(current)
		if currentT <= value {
			return // ok min-value written by other thread return
		}
	}
}

// Min returns current minimum value and a flag whether a value is present
//   - Thread-safe
func (a *AtomicMin[T]) Min() (value T, hasValue bool) {
	if hasValue = a.isInitialized.Load(); !hasValue {
		return // no min yet return
	}
	value = T(a.value.Load())
	return
}

// init uses lock to have loser threads wait until winner thread has updated value
func (a *AtomicMin[T]) init(valueU64 uint64) (didStore bool) {
	a.initLock.Lock()
	defer a.initLock.Unlock()

	if didStore = !a.isInitialized.Load(); !didStore {
		return // another thread was first
	}
	a.value.Store(valueU64)
	a.isInitialized.Store(true)
	return
}
