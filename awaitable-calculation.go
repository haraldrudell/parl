/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
)

// AwaitableCalculation contains an awaitable calculation using performant
// sync.RWMutex and atomics. Thread-safe
type AwaitableCalculation[T any] struct {
	// isCompleted makes lock atomic-access observable
	isCompleted AtomicBool

	// lock’s write lock is held by the calculating thread until the calculation completes
	//	- other threads are held waiting in Result using RLock for the calculation to complete
	lock sync.RWMutex
	// isValid indicates whether the calculation succeeded
	isValid bool
	// result holds any successful result of calculation
	//	- result value is the time calculating began
	//	- result holds zero-value if the calculation failed
	//	- result is updated by the winner thread prior to lock.Unlock
	result T // behind lock
}

// NewAwaitableCalculation returns an awaitable calculation using performant
// sync.RWMutex and atomics. Thread-safe
func NewAwaitableCalculation[T any]() (calculation *AwaitableCalculation[T]) {
	c := AwaitableCalculation[T]{}
	c.lock.Lock()
	return &c
}

// IsCompleted returns whether the calculation is complete. Thread-safe
func (cn *AwaitableCalculation[T]) IsCompleted() (isCompleted bool) {
	return cn.isCompleted.IsTrue()
}

// Result retrieves the calculation’s result. May block. Thread-safe
func (cn *AwaitableCalculation[T]) Result() (result T, isValid bool) {
	cn.lock.RLock()
	defer cn.lock.RUnlock()

	result = cn.result
	isValid = cn.isValid
	return
}

// End provides the result of the calculation
//   - result is considered valid if errp is nil or *errp is nil
func (cn *AwaitableCalculation[T]) End(result T, errp *error) {
	defer cn.lock.Unlock()
	defer cn.isCompleted.Set()

	// store value if calculation successful
	if cn.isValid = errp == nil || *errp == nil; cn.isValid {
		cn.result = result
	}
}
