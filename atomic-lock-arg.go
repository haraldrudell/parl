/*
© 2025-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

// AtomicLockArg provides a lazy-initialized singleton value
// behind atomics-shielded lock
//   - type parameter T: the type provided
//   - type parameter P: argument to creator function
//   - AtomicLockArg is AtomicLock with an argument for initialization
type AtomicLockArg[T any, P any] struct {
	// wrapped atomic lock providing hasT, lockT and t fields
	aLock AtomicLock[T]
}

// TMakerArg is a function initializing a T value pointed to
// that is provided a P-type argument
//   - tp: pointer to T field to initialize.
//   - — *tp is unreferenced zero-value
//   - arg: is an argument that was provided to [AtomicLockArg.Get].
//     arg allows TMakerArg to invoke methods on type P.
//     If arg was not provided to [AtomicLockArg.Get], arg is nil
//   - TMakerArg is invoked maximum once per AtomicLockArg
type TMakerArg[T any, P any] func(tp *T, arg *P)

// Get returns pointer to initialized T
//   - tMaker: function initializing a T value pointed to
//   - arg: optional argument to tMaker
//   - tp: pointer to initialized singleton T value
//   - —
//   - T can hold lock or atomic
//   - set hasT to true
//   - tMaker is invoked maximum once per AtomicLockArg
func (a *AtomicLockArg[T, P]) Get(tMaker TMakerArg[T, P], arg ...*P) (tp *T) {

	// tp’s pointer-value is known
	tp = &a.aLock.t

	// T already initialized case
	if a.aLock.hasT.Load() {
		return
	}

	// initialize T
	NilPanic("tMaker", tMaker)
	// get arg
	var p *P
	if len(arg) > 0 {
		p = arg[0]
	}
	// enter critical section to create T
	defer a.aLock.lockT.Lock().Unlock()

	// check inside lock
	if a.aLock.hasT.Load() {
		return
	}
	tMaker(tp, p)
	// t was created

	a.aLock.hasT.Store(true)

	return
}
