/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
	"time"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

// NBChan is a non-blocking send channel with trillion-size buffer.
//
//   - NBChan can be used as an error channel where the thread does not
//     block from a delayed or missing reader.
//   - errors can be read from the channel or fetched all at once using GetAll
//   - NBChan is initialization-free, thread-safe, idempotent, deferrable and observable.
//   - Ch(), Send(), Close() CloseNow() IsClosed() Count() are not blocked by channel send
//     and are panic-free.
//   - Close() CloseNow() are deferrable.
//   - WaitForClose() waits until the underlying channel has been closed.
//   - NBChan implements a thread-safe error store perrors.ParlError.
//   - NBChan.GetError() returns thread panics and close errors.
//   - No errors are added to the error store after the channel has closed.
//   - NBChan does not generate errors. When it does, errors are thread panics
//     or a close error. Neither is expected to occur
//   - the underlying channel is closed after Close is invoked and the channel is emptied
//   - cautious consumers may collect errors using the GetError method when:
//   - — the Ch receive-only channel is detected as being closed or
//   - — await using WaitForClose returns or
//   - — IsClosed method returns true
//
// Usage:
//
//	var errCh parl.NBChan[error]
//	go thread(&errCh)
//	err, ok := <-errCh.Ch()
//	errCh.WaitForClose()
//	errCh.GetError()
//	…
//	func thread(errCh *parl.NBChan[error]) {
//	defer errCh.Close() // non-blocking close effective on send complete
//	var err error
//	defer parl.Recover(parl.Annotation(), &err, errCh.AddErrorProc)
//	errCh.Ch() <- err // non-blocking
//	if err = someFunc(); err != nil {
//	err = perrors.Errorf("someFunc: %w", err)
//	return
type NBChan[T any] struct {
	closableChan ClosableChan[T]
	pendingSend  AtomicBool

	stateLock               sync.Mutex
	unsentCount             int  // inside lock
	sendQueue               []T  // inside lock
	waitForCloseInitialized bool // inside lock

	closeInvoked AtomicBool

	waitForClose sync.WaitGroup // initialization inside lock

	waitForThread WaitGroup // observable waitgroup

	perrors.ParlError // thread panics
}

// NewNBChan instantiates a non-blocking trillion-size buffer channel.
// NewNBChan allows initialization based on an existing channel.
// NewNBChan does not need initialization and can be used like:
//
//	var nbChan NBChan[error]
//	go thread(&nbChan)
func NewNBChan[T any](ch ...chan T) (nbChan *NBChan[T]) {
	nb := NBChan[T]{}
	nb.closableChan = *NewClosableChan(ch...) // store ch if present
	return &nb
}

// Ch obtains the receive-only channel
func (nb *NBChan[T]) Ch() (ch <-chan T) {
	return nb.closableChan.Ch()
}

// Send sends non-blocking on the channel
func (nb *NBChan[T]) Send(value T) {
	nb.stateLock.Lock()
	defer nb.stateLock.Unlock()

	if nb.closeInvoked.IsTrue() {
		return // no send after Close()
	}

	nb.unsentCount++

	// if no thread, send using new thread
	if nb.waitForThread.IsZero() {
		nb.pendingSend.Set()
		nb.waitForThread.Add(1)
		go nb.sendThread(value) // send err in new thread
		return
	}

	// put in queue
	nb.sendQueue = append(nb.sendQueue, value) // put err in send queue
}

// Get returns a slice of n or default all available items held by the channel.
func (nb *NBChan[T]) Get(n ...int) (allItems []T) {
	nb.stateLock.Lock()
	defer nb.stateLock.Unlock()

	// n0: 0 for all items, >0 that many items
	var n0 int
	if len(n) > 0 {
		n0 = n[0]
	}
	if n0 < 1 {
		n0 = 0
	}

	// get possible item from send thread
	var item T
	var itemValid bool
	for nb.pendingSend.IsTrue() && !itemValid {
		select {
		case item, itemValid = <-nb.closableChan.ch:
		default:
			time.Sleep(time.Millisecond)
		}
	}

	// allocate and populate allItems
	var itemLength int
	if itemValid {
		itemLength = 1
	}
	nq := len(nb.sendQueue)
	// cap n to set n0
	if n0 != 0 && nq > n0 {
		nq = n0
	}
	allItems = make([]T, nq+itemLength)
	if itemValid {
		allItems[0] = item
	}

	// empty the send buffer
	if nq > 0 {
		copy(allItems[itemLength:], nb.sendQueue)
		copy(nb.sendQueue, nb.sendQueue[nq:])
		nb.sendQueue = nb.sendQueue[:len(nb.sendQueue)-nq]
		nb.unsentCount -= nq
	}

	return
}

// Count returns number of unsent values
func (nb *NBChan[T]) Count() (unsentCount int) {
	nb.stateLock.Lock()
	defer nb.stateLock.Unlock()

	return nb.unsentCount
}

// Close orders the channel to close once pending sends complete.
// Close is thread-safe, non-blocking and panic-free.
func (nb *NBChan[T]) Close() (didClose bool) {
	nb.stateLock.Lock()
	defer nb.stateLock.Unlock()

	if !nb.closeInvoked.Set() {
		return // Close was already invoked
	}

	if !nb.waitForThread.IsZero() {
		return // there is a pending thread that will execute close on exit
	}
	var err error
	if didClose, err = nb.closableChan.Close(); didClose && err != nil { // execute the close now
		nb.AddError(err) // store posible close error
	}
	return
}

func (nb *NBChan[T]) DidClose() (didClose bool) {
	return nb.closeInvoked.IsTrue()
}

// IsClosed indicates whether the channel has actually closed.
func (nb *NBChan[T]) IsClosed() (isClosed bool) {
	return nb.closableChan.IsClosed()
}

func (nb *NBChan[T]) WaitForClose() {
	nb.initWaitForClose() // ensure waitForClose state is valid
	nb.waitForClose.Wait()
}

// CloseNow closes without waiting for sends to complete.
// Close does not panic.
// Close is thread-safe.
// Close does not return until the channel is closed.
// Upon return, all invocations have a possible close error in err.
// if errp is non-nil, it is updated with error status
func (nb *NBChan[T]) CloseNow(errp ...*error) (err error, didClose bool) {
	if nb.closableChan.IsClosed() {
		return // channel is already closed
	}
	nb.stateLock.Lock()
	defer nb.stateLock.Unlock()

	nb.closeInvoked.Set()

	// discard pending data
	if len(nb.sendQueue) > 0 {
		nb.sendQueue = nil
		nb.unsentCount = 0
	}

	// close the channel now
	if didClose, err = nb.closableChan.Close(); didClose && err != nil { // execute the close now
		nb.AddError(err) // store posible close error
	}
	return
}

func (nb *NBChan[T]) sendThread(value T) {
	defer nb.sendThreadDefer()
	defer Recover(Annotation(), nil, func(err error) {
		if pruntime.IsSendOnClosedChannel(err) {
			return // ignore if the channel was or became closed
		}
		nb.AddError(err)
	})
	defer nb.pendingSend.Clear()

	ch := nb.closableChan.Ch()
	for {
		ch <- value // may block or panic
		nb.pendingSend.Clear()

		var ok bool
		if value, ok = nb.valueToSend(); !ok {
			break
		}
	}
}

func (nb *NBChan[T]) valueToSend() (value T, ok bool) {
	nb.stateLock.Lock()
	defer nb.stateLock.Unlock()

	// count the item ust sent
	nb.unsentCount--

	// no more values: end thread
	if len(nb.sendQueue) == 0 {
		return
	}

	// send next value in queue
	value = nb.sendQueue[0]
	ok = true
	copy(nb.sendQueue[0:], nb.sendQueue[1:])
	nb.sendQueue = nb.sendQueue[:len(nb.sendQueue)-1]
	nb.pendingSend.Set()
	return
}

func (nb *NBChan[T]) sendThreadDefer() {
	if nb.closeInvoked.IsTrue() { // Close() was invoked after thread started
		nb.closableChan.Close() // close if Close was invoked. Idempotent
	}

	nb.waitForThread.Done() // thread has exit
}

func (nb *NBChan[T]) initWaitForClose() {
	nb.stateLock.Lock()
	defer nb.stateLock.Unlock()

	if nb.waitForCloseInitialized {
		return // state is valid
	}
	nb.waitForCloseInitialized = true

	if !nb.closableChan.IsClosed() {
		nb.waitForClose.Add(1) // has to wait for close to occur
	}
}
