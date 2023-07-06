/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"math"
	"sync"
	"sync/atomic"
)

const (
	lockAccessBit       = uint64(1)
	ticketDelta         = uint64(2)
	negativeTicketDelta = math.MaxUint64 - ticketDelta + 1
	atomicsComplete     = lockAccessBit
)

// accessManager controls atomic and lock-based access
//   - Lock access: defer a.Lock().Unlock()
//   - atomic or lock access: defer RelinquishAccess(RequestAccess())
type accessManager struct {
	// lock for data access used during lock access periods
	accessLock sync.Mutex

	// number of requestors requiring lock use, atomic
	lockers uint64
	// lock for changing atomic access status
	//	- ensures sequential by-single-thread transitions of lockAccessBit
	//	- reduces [accessManager.before] contention
	controlLock sync.Mutex
	// atomic access ongoing operations counter, atomic
	//	- lowest bit is lockAccessBit
	//	- lockAccessBit set and cleared behind controlLock
	before uint64
}

// RequestAccess allows a data provider to make writes to
// shared data without knowing if access is atomic or behind lock
//
// Usage:
//
//	func f() {
//	  defer RelinquishAccess(RequestAccess())
//	  …
func (a *accessManager) RequestAccess() (isLockAccess bool) {

	// deteremine access method
	for {
		var before = atomic.LoadUint64(&a.before)
		isLockAccess = before&lockAccessBit != 0
		// attempt atomic access
		if !isLockAccess {
			if atomic.CompareAndSwapUint64(&a.before, before, before+ticketDelta) {
				return // atomic access return: isLockAccess: false
			}
			continue // try again
		}

		// it is lock access
		// acquire access lock
		a.accessLock.Lock()
		return // lock acquired return: isLockAccess true
	}
}

// RelinquishAccess signals that a provider is done writing to
// shared data
func (a *accessManager) RelinquishAccess(isLockAccess bool) {
	if isLockAccess {
		// end lock access
		a.accessLock.Unlock()
		return
	}

	// end atomic access
	atomic.AddUint64(&a.before, negativeTicketDelta)
}

// Lock sets access behind lock, clears out atomic operations and
// acquires the access-lock
//
// Usage:
//
//	func f() {
//	  defer a.Lock().Unlock()
//	  …
func (a *accessManager) Lock() (a2 *accessManager) {
	a2 = a
	var before = a.disableAtomic()
	// ensure atomic operations ceased
	for before != atomicsComplete {
		before = atomic.LoadUint64(&a.before)
	}
	// acquire lock
	a.accessLock.Lock()
	return
}

// Unlock releases the access-lock
func (a *accessManager) Unlock() {
	a.enableAtomic()
	a.accessLock.Unlock()
}

// disableAtomic transitions to lock-access
//   - on return lock access is set and atomic operations have ceased
func (a *accessManager) disableAtomic() (before uint64) {

	// prevent lockRequestBit from being cleared
	//	- a.lockers > 0
	atomic.AddUint64(&a.lockers, 1)
	//	- check current state
	if before = atomic.LoadUint64(&a.before); before&lockAccessBit != 0 {
		return // bit set but atomic operations may be ongoing return
	}
	a.controlLock.Lock()
	defer a.controlLock.Unlock()

	// ensure lockAccessBit true
	//	- check current state
	if before = atomic.LoadUint64(&a.before); before&lockAccessBit != 0 {
		return // bit set but atomic operations may be ongoing return
	}

	// set lockAccessBit
	for {
		var oldBefore = atomic.LoadUint64(&a.before)
		before = oldBefore | lockAccessBit
		if atomic.CompareAndSwapUint64(&a.before, oldBefore, before) {
			return // bit set but atomic operations may be ongoing return
		}
	}
}

// enableAtomic ends a lock-access period which may re-enable atomic access
func (a *accessManager) enableAtomic() {

	// decrement and check if atomic access should be re-enabled
	if atomic.AddUint64(&a.lockers, ^uint64(0)) > 0 {
		return // do not re-enable atomic access: pending lock-use requestors return
	}
	// check if bit already cleared
	if atomic.LoadUint64(&a.before)&lockAccessBit == 0 {
		return // bit already cleared return
	}
	a.controlLock.Lock()
	defer a.controlLock.Unlock()

	// clear lock-access bit
	// check if bit already cleared
	if atomic.LoadUint64(&a.before)&lockAccessBit == 0 {
		return // bit already cleared return
	}
	for {

		// if pending Lock invocation: abort
		if atomic.LoadUint64(&a.lockers) > 0 {
			return // no re-enable: more requestors return
		}

		// attempt to clear bit
		var before = atomic.LoadUint64(&a.before)
		if atomic.CompareAndSwapUint64(&a.before, before, before&^lockAccessBit) {
			return // bit was cleared return
		}
	}
}
