/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "sync"

// MutexWait is maximum-lightweight observable single-fire Mutex. Thread-Safe
type MutexWait struct {
	lock       sync.Mutex
	isUnlocked AtomicBool
}

// NewMutexWait returns a maximum-lightweight observable single-fire Mutex. Thread-Safe
func NewMutexWait() (mutexWait *MutexWait) {
	mutexWait = &MutexWait{}
	mutexWait.lock.Lock()
	return
}

// IsUnlocked returns whether the MutexWait has fired
func (mw *MutexWait) IsUnlocked() (isUnlocked bool) {
	return mw.isUnlocked.IsTrue()
}

// Wait blocks until MutexWait has fired
func (mw *MutexWait) Wait() {
	mw.lock.Lock()
	defer mw.lock.Unlock()
}

// Unlock fires MutexWait
func (mw *MutexWait) Unlock() {
	if mw.isUnlocked.Set() {
		mw.lock.Unlock()
	}
}
