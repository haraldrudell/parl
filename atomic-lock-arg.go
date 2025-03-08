/*
© 2025-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

// AtomicLockArg provides a lazy-initialized singleton value
// behind atomics-shielded lock
//   - T is type created
//   - P is argument to creator function
//   - AtomicLockArg is used when creating T requires a parameter
type AtomicLockArg[T any, P any] struct {
	aLock AtomicLock[T]
}

// TMakerArg is creator function that is provided argument
//   - tp: where to create T, *tp is zero-value
//   - arg: argument provided to [AtomicLockArg.Get].
//     arg allows for invoking method on struct value
//   - invoked once per AtomicLockArg
type TMakerArg[T any, P any] func(tp *T, arg *P)

// Get returns T possibly creating it using tMaker
//   - tMaker: creator function
//   - arg: argument to tMaker
//   - tp: pointer to created singleton T
//   - T can hold lock or atomic
func (a *AtomicLockArg[T, P]) Get(tMaker TMakerArg[T, P], arg ...*P) (tp *T) {

	// T already created case
	if a.aLock.hasT.Load() {
		tp = &a.aLock.t
		return
	}

	NilPanic("tMaker", tMaker)
	var p *P
	if len(arg) > 0 {
		p = arg[0]
	}
	// enter critical section to create T
	defer a.aLock.lockT.Lock().Unlock()

	// check inside lock
	if a.aLock.hasT.Load() {
		tp = &a.aLock.t
		return
	}
	tMaker(&a.aLock.t, p)
	// t was created

	tp = &a.aLock.t
	a.aLock.hasT.Store(true)

	return
}
