/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

import (
	"fmt"
	"os"
	"sync"
)

const peNil = "<nil>"

// ParlError is a thread-safe error container
type ParlError struct {
	e error // inside lock
	sync.RWMutex
	errCh chan<- error // inside lock
}

var _ error = &ParlError{}      // ParlError behaves like an error
var _ ErrorStore = &ParlError{} // ParlError is an error store

func NewParlError(errCh chan<- error) (pe *ParlError) {
	return &ParlError{errCh: errCh}
}

// AddError stores additional errors in the comtainer.
// It is thread-safe and returns the current state.
// For a non-thread-safe version, use error116.Errp
func (pe *ParlError) AddError(err error) (e error) {
	if err == nil {
		return pe.GetError()
	}
	pe.RWMutex.Lock()
	defer pe.RWMutex.Unlock()
	if pe.errCh != nil {
		pe.errCh <- err
	}
	if pe.e == nil {
		pe.e = err
	} else {
		pe.e = AppendError(pe.e, err)
	}
	return pe.e
}

// AddErrorProc stores additional errors in the container
// It is thread-safe and has a no-return-value signature.
// For a non-thread-safe version, use error116.Errp
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

// Shutdown is only required if this ParlError has an error channel
func (pe *ParlError) Shutdown() {
	pe.RWMutex.Lock()
	defer pe.RWMutex.Unlock()
	defer func() {
		if v := recover(); v != nil {
			// This will never happen. This is the best we can do when it does
			err, ok := v.(error)
			if !ok {
				err = Errorf("ParlError panic on closing errCh: %+v", v)
			} else {
				err = Stack(err)
			}
			fmt.Fprintln(os.Stderr, err)
			pe.AddError(err)
		}
	}()

	if pe.errCh == nil {
		return // there is no channel to close
	}
	ec := pe.errCh
	pe.errCh = nil // prevents further send and close
	close(ec)
}
