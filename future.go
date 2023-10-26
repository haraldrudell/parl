/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
	"sync/atomic"
)

// Future contains an awaitable calculation using performant
// sync.RWMutex and atomics. Thread-safe
type Future[T any] struct {
	// isCompleted makes lock atomic-access observable
	isCompleted atomic.Bool

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

// NewFuture returns an awaitable calculation using performant
// sync.RWMutex and atomics. Thread-safe
func NewFuture[T any]() (calculation *Future[T]) {
	c := Future[T]{}
	c.lock.Lock()
	return &c
}

// IsCompleted returns whether the calculation is complete. Thread-safe
func (cn *Future[T]) IsCompleted() (isCompleted bool) {
	return cn.isCompleted.Load()
}

// Result retrieves the calculation’s result. May block. Thread-safe
func (cn *Future[T]) Result() (result T, isValid bool) {
	cn.lock.RLock() // invokers block here
	defer cn.lock.RUnlock()

	result = cn.result
	isValid = cn.isValid
	return
}

// End writes the result of the calculation, deferrable
//   - result is considered valid if errp is nil or *errp is nil
func (cn *Future[T]) End(result *T, errp *error) {
	defer cn.lock.Unlock()
	defer cn.isCompleted.Store(true)

	// store value if calculation successful
	if cn.isValid = errp == nil || *errp == nil; cn.isValid {
		cn.result = *result
	}
}
