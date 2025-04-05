/*
© 2025-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import (
	"sync/atomic"
)

// AtomicLock provides a lazy-initialized singleton value
// behind atomics-shielded lock
//   - type parameter T: the type of the value being lazily created
//   - AtomicLock is used when T is created without state
type AtomicLock[T any] struct {
	// hasT is true once T is available
	//	- allows for thread-safe channel value read outside lock
	//		at atomic performance
	hasT atomic.Bool
	// lock protects T write
	//	- lock is only used for initial write while hasT is false
	lockT Mutex
	// T singleton value
	//	- directly readable whenever hasT observed true
	//	- lazy initialization
	//	- may contain lock or atomic
	t T
}

// TCreator is object creating T value
type TCreator[T any] interface {
	// MakeT initializes T value at *tp
	//   - *tp: uninitialized T zero-value
	//   - —
	//   - tp points to a T field that has not been referenced
	//   - MakeT is invoked maximum once per T field
	MakeT(tp *T)
}

// TMaker is function initializing T value at *tp
//   - *tp: uninitialized T zero-value
//   - —
//   - tp points to a T field that has not been referenced
//   - TMaker is invoked maximum once per T field
type TMaker[T any] func(tp *T)

// Get returns T value possibly initializing it using [tCreator.MakeT]
//   - tCreator: object able to initialize T values
//   - tp: points to singleton value
//   - —
//   - T can hold lock or atomic
//   - MakeT is invoked maximum once per AtomicLock
func (a *AtomicLock[T]) Get(tCreator TCreator[T]) (tp *T) {

	// tp value is known
	tp = &a.t

	// T already created case
	if a.hasT.Load() {
		return
	}

	NilPanic("tCreator", tCreator)
	a.get(nil, tCreator)

	return
}

// GetFunc returns T value possibly initializing it using tMaker function
//   - tMaker: function initializing a T value pointed to
//   - tp: points to initialized singleton value
//   - —
//   - T can hold lock or atomic
//   - tMaker is invoked maximum once per AtomicLock
func (a *AtomicLock[T]) GetFunc(tMaker TMaker[T]) (tp *T) {

	// tp value is known
	tp = &a.t

	// T already created case
	//	- atomic performance
	if a.hasT.Load() {
		return
	}

	// initialize a.t
	NilPanic("tMaker", tMaker)
	a.get(tMaker, nil)

	return
}

// get enters critical section to get or create T
//   - tMaker: function initializing a T value pointed to
//   - tCreator.MakeT: method initializing a T value pointed to
//   - tp: pointer to initialized T value
//     -
//   - sets a.hasT to true
//   - tMaker or tCreator: one must be non-nil, tMaker is checked first
//   - tp: pointer to T
func (a *AtomicLock[T]) get(tMaker TMaker[T], tCreator TCreator[T]) {
	defer a.lockT.Lock().Unlock()

	// check inside lock
	if a.hasT.Load() {
		return
	}

	// initialize a.t
	if tMaker != nil {
		tMaker(&a.t)
	} else {
		tCreator.MakeT(&a.t)
	}
	// t was initialized

	a.hasT.Store(true)
}
