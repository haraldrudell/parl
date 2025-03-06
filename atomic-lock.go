/*
© 2025-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import (
	"sync/atomic"
)

// AtomicLock provides a lazy-initialized value behind atomics-shielded lock
type AtomicLock[T any] struct {
	// hasChan is true once T is available
	//	- allows for thread-safe channel value read outside lock
	//		at atomic performance
	hasT atomic.Bool
	// lock protects ch write
	//	- lock is only used for the initial write
	lockT Mutex
	// channel written behind lock
	//	- directly readable whenever hasChan is observed true
	//	- lazy initialization
	t T
}

// TCreator is object creating T value
type TCreator[T any] interface {
	// MakeT creates T value at tp
	MakeT(tp *T)
}

// TMaker is a function creating T value at tp
type TMaker[T any] func(tp *T)

// Get returns T value possibly creating it using tCreator
func (a *AtomicLock[T]) Get(tCreator TCreator[T]) (tp *T) {

	// T already created case
	if a.hasT.Load() {
		tp = &a.t
		return
	}

	NilPanic("tCreator", tCreator)
	return a.get(nil, tCreator)
}

// Get returns T value possibly creating it using tMaker
func (a *AtomicLock[T]) GetFunc(tMaker TMaker[T]) (tp *T) {

	// T already created case
	if a.hasT.Load() {
		tp = &a.t
		return
	}

	NilPanic("tMaker", tMaker)
	return a.get(tMaker, nil)
}

// get enters critical section to get or create T
//   - tMaker or tCreator: one must be non-nil
//   - tp: pointer to T
func (a *AtomicLock[T]) get(tMaker TMaker[T], tCreator TCreator[T]) (tp *T) {
	defer a.lockT.Lock().Unlock()

	// check inside lock
	if a.hasT.Load() {
		tp = &a.t
		return
	} else if tMaker != nil {
		tMaker(&a.t)
	} else {
		tCreator.MakeT(&a.t)
	}
	// t was created

	tp = &a.t
	a.hasT.Store(true)

	return
}
