/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

import (
	"sync"
)

const peNil = "<nil>"

// ParlError is a thread-safe error container
type ParlError struct {
	e error
	sync.RWMutex
}

var _ error = &ParlError{} // ParlError behaves like an error

// AddError stores additional errors in the comtainer. Thread-safe.
// for a non-thread-safe version, use error116.Errp
func (pe *ParlError) AddError(err error) (e error) {
	if err == nil {
		return pe.GetError()
	}
	pe.RWMutex.Lock()
	defer pe.RWMutex.Unlock()
	if pe.e == nil {
		pe.e = err
	} else {
		pe.e = AppendError(pe.e, err)
	}
	return pe.e
}

func (pe *ParlError) AddErrorProc(err error) {
	pe.AddError(err)
}

// GetError returns the error value enclosed in ParlError. Thread-safe
func (pe *ParlError) GetError() (e error) {
	pe.RWMutex.RLock()
	defer pe.RWMutex.RUnlock()
	return pe.e
}

// Error() makes ParlError behave like an error.
// Error is thread-safe unlike in most other Error implementations.
// Because code will check if ParlError is nil, which it mostly isn’t, and then
// invoke .Error(), it may be that Error is invoked when the error field is nil.
// For those situations, return “<nil>” like fmt.Print might do
func (pe *ParlError) Error() string {
	if err := pe.GetError(); err != nil {
		return err.Error()
	}
	return peNil
}
