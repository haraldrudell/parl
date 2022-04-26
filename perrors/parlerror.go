/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"sync"

	"github.com/haraldrudell/parl/errorglue"
)

const (
	peNil = "<nil>"
)

// ParlError is a thread-safe error container that can optionally
// send errors non-blocking on a channel.
type ParlError struct {
	errLock          sync.RWMutex
	err              error // inside lock
	errorglue.SendNb       // non-blocking channel for sending errors
}

var _ error = &ParlError{}                // ParlError behaves like an error
var _ errorglue.ErrorStore = &ParlError{} // ParlError is an error store

/*
NewParlError provides a thread-safe error container that can optionally
send incoming errors non-blocking on a channel.

If a channel is not used, a zero-value works:
 var err error116.ParlError
 …
 return err

When using a channel, The error channel is closed by Shutdown():
 errCh := make(chan error)
 err := NewParlError(errCh)
 …
 err.Shutdown()
 …
 if err, ok := <- errCh; !ok {
   // errs was shutdown
A shutdown ParlError is still usable, but will no longer send errors
*/
func NewParlError(errCh chan<- error) (pe *ParlError) {
	p := ParlError{}
	if errCh != nil {
		p.SendNb.SendChannel = *errorglue.NewSendChannel(errCh)
	}
	return &p
}

// AddError stores additional errors in the container.
// Thread-safe. Returns the current state.
// For a non-thread-safe version, use error116.Errp
func (pe *ParlError) AddError(err error) (err1 error) {

	// if there is no error, no change
	if err == nil {
		return pe.GetError()
	}

	// update the container
	err1 = pe.addError(err)

	// send if we have a channel, never block
	pe.Send(err)

	return
}

// AddErrorProc stores additional errors in the container
// It is thread-safe and has a no-return-value signature.
// For a non-thread-safe version, use error116.Errp
func (pe *ParlError) AddErrorProc(err error) {
	pe.AddError(err)
}

// GetError returns the error value enclosed in ParlError. Thread-safe
func (pe *ParlError) GetError() (err error) {
	pe.errLock.RLock()
	defer pe.errLock.RUnlock()
	return pe.err
}

// InvokeIfError invokes fn if the error store contains an error
func (pe *ParlError) InvokeIfError(fn func(err error)) {
	if err := pe.GetError(); err != nil {
		fn(err)
	}
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

func (pe *ParlError) addError(err error) (err1 error) {
	pe.errLock.Lock()
	defer pe.errLock.Unlock()

	// determine the new value
	if pe.err == nil {
		err1 = err
	} else {
		err1 = AppendError(pe.err, err)
	}

	// store the new value
	pe.err = err1

	return
}
