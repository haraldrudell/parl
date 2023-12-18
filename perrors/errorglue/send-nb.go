/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

// SendNb implements non-blocking send using a thread and buffer up to size of int
type SendNb struct {
	errCh chan error // may be nil: check using method HasChannel

	sqLock    sync.Mutex
	sendQueue []error       // inside lock
	hasThread bool          // inside lock
	waitCh    chan struct{} // before go statement

	isShutdown atomic.Bool // atomic: check using method: IsShutdown
}

// NewSendNb returns a buffered Send object using errCh
//   - errCh may be nil
func NewSendNb(errCh chan error) (sc *SendNb) {
	return &SendNb{errCh: errCh}
}

// Send sends an error on the error channel. Non-blocking. Thread-safe.
//   - if err is nil, nothing is done
//   - if SendNb was not initialized with non-zero channel, nothing is done
func (sc *SendNb) Send(err error) {
	if !sc.canSend(err) {
		return // no-data not-initialized is-shutdown return
	}

	sc.sqLock.Lock()
	defer sc.sqLock.Unlock()

	if !sc.canSend(err) {
		return // no-data not-initialized is-shutdown return
	}

	// put err in send queue
	sc.sendQueue = append(sc.sendQueue, err)
	if sc.hasThread {
		return // thread is already reading from queue
	}

	// launch thread
	sc.hasThread = true
	sc.waitCh = make(chan struct{})
	go sc.sendThread() // send err in new thread
}

func (sc *SendNb) HasChannel() (hasChannel bool) {
	return sc.errCh != nil
}

func (sc *SendNb) IsShutdown() (isShutdown bool) {
	return sc.isShutdown.Load()
}

// Shutdown closes the channel exactly once. Thread-safe
func (sc *SendNb) Shutdown() {
	if sc.isShutdown.Load() {
		return // already shutdown
	}

	// ensure channel completes send
	if sc.shutdown() {
		select {
		case <-sc.errCh: // release thread from send wait
			<-sc.waitCh // wait for thread to exit
		case <-sc.waitCh: // or wait until thread has exited
		}
	}

	if sc.errCh != nil {
		close(sc.errCh)
	}
}

func (sc *SendNb) shutdown() (hasThread bool) {
	sc.sqLock.Lock()
	defer sc.sqLock.Unlock()

	// block further Send invocation and thread sends
	if sc.isShutdown.Swap(true) {
		return // already shutdown
	}

	hasThread = sc.hasThread

	// drop any send queue
	sc.sendQueue = nil

	return
}

func (sc *SendNb) canSend(err error) (canSend bool) {
	return err != nil && // nothing to send
		sc.errCh != nil && // not initialized to send errors
		!sc.isShutdown.Load() // is shutdown
}

func (sc *SendNb) sendThread() {
	defer sc.closeWaitCh()
	defer RecoverThread("send on error channel", func(err error) {
		fmt.Fprintln(os.Stderr, err)
	})

	var hasValue bool
	var err error
	for {
		if hasValue, err = sc.getErrFromSendQueue(); !hasValue {
			return // queue empty return: exit thread
		}

		sc.errCh <- err
	}
}

func (sc *SendNb) getErrFromSendQueue() (hasValue bool, err error) {
	sc.sqLock.Lock()
	defer sc.sqLock.Unlock()

	if hasValue = len(sc.sendQueue) > 0; !hasValue {
		sc.hasThread = false
		return
	}

	err = sc.sendQueue[0]
	copy(sc.sendQueue, sc.sendQueue[1:])
	lastIndex := len(sc.sendQueue) - 1
	sc.sendQueue[lastIndex] = nil
	sc.sendQueue = sc.sendQueue[:lastIndex]
	return
}

func (sc *SendNb) closeWaitCh() {
	sc.sqLock.Lock()
	defer sc.sqLock.Unlock()

	close(sc.waitCh)
	sc.hasThread = false
}
