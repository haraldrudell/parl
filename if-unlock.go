/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// Unlocker facilitates one-liner Lock-Unlock
type Unlocker interface {
	// Unlock unlocks m.
	// It is a run-time error if m is not locked on entry to Unlock.
	//
	// A locked [Mutex] is not associated with a particular goroutine.
	// It is allowed for one goroutine to lock a Mutex and then
	// arrange for another goroutine to unlock it.
	Unlock()
}

// RUnlocker facilitates one-liner RLock-RUnlock
type RUnlocker interface {
	// RUnlock undoes a single [RWMutex.RLock] call;
	// it does not affect other simultaneous readers.
	// It is a run-time error if rw is not locked for reading
	// on entry to RUnlock.
	RUnlock()
}

// LockUnLock can lock and unlock resources
type LockUnlock interface {
	// Lock locks m.
	// If the lock is already in use, the calling goroutine
	// blocks until the mutex is available.
	Lock() (unlocker Unlocker)
}
