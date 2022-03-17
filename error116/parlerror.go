/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

import "sync"

// ParlError is a thread-safe error container
type ParlError struct {
	error
	sync.RWMutex
}

// Add stores additional errors in the comtainer. Thread-safe
func (pe *ParlError) Add(err error) (e error) {
	if err == nil {
		return pe.Get()
	}
	pe.RWMutex.Lock()
	defer pe.RWMutex.Unlock()
	if pe.error == nil {
		pe.error = err
	} else {
		pe.error = AppendError(pe.error, err)
	}
	return pe.error
}

// Get returns the error value enclosed in ParlError. Thread-safe
func (pe *ParlError) Get() (e error) {
	pe.RWMutex.RLock()
	defer pe.RWMutex.RUnlock()
	return pe.error
}
