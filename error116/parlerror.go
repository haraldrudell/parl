/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
)

const (
	peNil                 = "<nil>"
	peSendOnClosedChannel = "send on closed channel"
)

// ParlError is a thread-safe error container
type ParlError struct {
	errLock   sync.RWMutex
	err       error // inside lock
	errChLock sync.Mutex
	errCh     chan<- error // value and close inside lock
}

var _ error = &ParlError{}      // ParlError behaves like an error
var _ ErrorStore = &ParlError{} // ParlError is an error store

/*
NewParlError sends its errors on the provided error channel in addition to storing
them in the error container. Thread safe. The error channel is closed by
(*ParlError).Shutdown().
For ParlError instances without error channel, use a composite initializer or
or provide nil value
*/
func NewParlError(errCh chan<- error) (pe *ParlError) {
	return &ParlError{errCh: errCh}
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
	errCh := pe.getErrCh()
	if errCh == nil {
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go pe.send(errCh, err1, &wg) // non-blocking send
	wg.Wait()                    // wait for thread send to get ready

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
	defer pe.recover("ParlError panic on closing errCh", func(err error) {
		// This will never happen. This is the best we can do when it does
		fmt.Fprintln(os.Stderr, err)
		pe.AddError(err)
	})

	pe.closeAction() // thread-safe, closes exactly once
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

func (pe *ParlError) getErrCh() (errCh chan<- error) {
	pe.errChLock.Lock()
	defer pe.errChLock.Unlock()
	return pe.errCh
}

func (pe *ParlError) closeAction() {
	pe.errChLock.Lock()
	defer pe.errChLock.Unlock()
	errCh := pe.errCh
	if errCh == nil {
		return
	}
	pe.errCh = nil // prevent further send or close
	close(errCh)
}

func (pe *ParlError) recover(label string, onError func(err error)) {
	if v := recover(); v != nil {
		err, ok := v.(error)
		if !ok {
			err = Errorf("ParlError panic on closing errCh: %+v", v)
		} else {
			err = Errorf("%s: %w", label, err)
		}
		onError(err)
	}
}

func (pe *ParlError) send(errCh chan<- error, err error, wg *sync.WaitGroup) {
	defer pe.recover("send on error channel", func(err error) {
		if isSendOnClosedChannel(err) {
			return // ignore if the channel was or became closed
		}
		pe.addError(err) // add to the lot
	})

	wg.Done()    // signal read-to-send
	errCh <- err // may block and panic
}

func isSendOnClosedChannel(err error) (is bool) {

	// errors.As cannot handle nil
	if err == nil {
		return false // not an error
	}

	// is it a runtime error?
	// Go1.18: err value is a private custom type string
	// the value is effectively inaccessible
	// runtime defines an interface for its errors
	var runtimeError runtime.Error
	if !errors.As(err, &runtimeError) {
		return false // not a runtime error
	}

	// is it the right runtime error?
	return runtimeError.Error() == peSendOnClosedChannel
}
