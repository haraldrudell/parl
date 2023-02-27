/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
	"github.com/haraldrudell/parl/pslices"
)

// NBChan is a non-blocking send channel with trillion-size buffer.
//   - NBChan behaves both like a channel and a thread-safe slice
//   - NBChan has non-blocking, thread-safe, error-free and panic-free Send and SendMany
//   - NBChan has deferrable, panic-free, idempotent close
//   - NBChan is initialization-free, thread-safe, panic-free, idempotent, deferrable and observable.
//   - NBChan can be used as an error channel where the sending thread does not
//     block from a delayed or missing reader.
//   - errors can be read from the channel or fetched all at once using GetAll
//   - Ch(), Send(), Close() CloseNow() IsClosed() Count() are not blocked by channel send
//     and are panic-free.
//   - values are sent using Send or SendMany methods
//   - values are read from Ch channel or using Get method
//   - Close, CloseNow and WaitForClose are deferrable.
//   - WaitForClose waits until the underlying channel has been closed.
//   - NBChan implements a thread-safe error store perrors.ParlError.
//   - NBChan.GetError() returns thread panics and close errors.
//   - No errors are added to the error store after the channel has closed.
//   - NBChan’s only errors are thread panics and close errors.
//     Neither are expected to occur
//   - the underlying channel is closed after Close is invoked and the channel is emptied
//   - cautious consumers may collect errors via:
//   - — CloseNow or WaitForClose
//   - — GetError method preferrably after CloseNow, WaitForClose or IsClosed returns true
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

	stateLock       sync.Mutex
	unsentCount     int  // inside lock. One item may be with sendThread
	sendQueue       []T  // inside lock. One item may be with sendThread
	isRunningThread bool // inside lock
	// closesOnThreadSend is created inside the lock every time a value is provided
	// to the send thread
	//	- closesOnThreadSend will close immediately after the thread sends
	//	- from inside the lock, this allows for the thread’s value to be collected
	closesOnThreadSend        chan struct{} // write inside lock
	isCloseInvoked            AtomicBool    // Set inside lock
	isWaitForCloseInitialized AtomicBool    // Set inside lock

	waitForClose sync.WaitGroup // valid when isWaitForCloseInitialized is true

	perrors.ParlError // thread panics and close errors
}

// NewNBChan returns a non-blocking trillion-size buffer channel.
//   - NewNBChan allows initialization based on an existing channel.
//   - NBChan does not need initialization and can be used like:
//
// Usage:
//
//	var nbChan NBChan[error]
//	go thread(&nbChan)
func NewNBChan[T any](ch ...chan T) (nbChan *NBChan[T]) {
	nb := NBChan[T]{}
	if len(ch) > 0 {
		nb.closableChan = *NewClosableChan(ch[0]) // store ch if present
	}
	return &nb
}

// Ch obtains the receive-only channel
func (nb *NBChan[T]) Ch() (ch <-chan T) {
	return nb.closableChan.Ch()
}

// Send sends a single value non-blocking, thread-safe, panic-free and error-free on the channel
func (nb *NBChan[T]) Send(value T) {
	if nb.isCloseInvoked.IsTrue() {
		return // no send after Close(), atomic performance
	}
	nb.stateLock.Lock()
	defer nb.stateLock.Unlock()

	if nb.isCloseInvoked.IsTrue() {
		return // no send after Close()
	}

	nb.unsentCount++

	// if thread is running, append to send queue
	if nb.isRunningThread {
		nb.sendQueue = append(nb.sendQueue, value)
		return
	}

	// send using new thread
	nb.startThread(value)
}

// Send sends many values non-blocking, thread-safe, panic-free and error-free on the channel
func (nb *NBChan[T]) SendMany(values []T) {
	if nb.isCloseInvoked.IsTrue() {
		return // no send after Close(), atomic performance
	}
	var length int
	if length = len(values); length == 0 {
		return // nothing to do return
	}
	nb.stateLock.Lock()
	defer nb.stateLock.Unlock()

	if nb.isCloseInvoked.IsTrue() {
		return // no send after Close()
	}

	nb.unsentCount += length

	// if thread is running, add to send queue
	if nb.isRunningThread {
		nb.sendQueue = append(nb.sendQueue, values...)
		return
	}

	// get next value to send, append remainign to send queue
	var value T
	if len(nb.sendQueue) > 0 {
		value = nb.sendQueue[0]
		pslices.TrimLeft(&nb.sendQueue, 1)
		nb.sendQueue = append(nb.sendQueue, values...)
	} else {
		value = values[0]
		nb.sendQueue = append(nb.sendQueue, values[1:]...)
	}

	nb.startThread(value)
}

// Get returns a slice of n or default all available items held by the channel.
//   - if channel is empty, 0 items are returned
//   - Get is non-blocking
//   - n > 0: max this many items
//   - n == 0 (or <0): all items
//   - Get is panic-free non-blocking error-free thread-safe
func (nb *NBChan[T]) Get(n ...int) (allItems []T) {
	nb.stateLock.Lock()
	defer nb.stateLock.Unlock()

	// n0: 0 for all items, >0 that many items
	var n0 int
	if len(n) > 0 {
		n0 = n[0]
	}
	if n0 < 0 {
		n0 = 0
	}

	// get possible item from send thread
	var item T
	var itemValid bool
	if nb.isRunningThread {
		select {
		case <-nb.closesOnThreadSend:
		case item, itemValid = <-nb.closableChan.ch:
		}
	}

	// allocate and populate allItems
	var itemLength int
	if itemValid {
		itemLength = 1
	}
	nq := len(nb.sendQueue)
	// cap n to set n0
	if n0 != 0 && nq+itemLength > n0 {
		nq = n0 - itemLength
	}
	allItems = make([]T, nq+itemLength)
	// possible item from channel
	if itemValid {
		allItems[0] = item
	}
	// items from sendQueue
	if nq > 0 {
		copy(allItems[itemLength:], nb.sendQueue)
		pslices.TrimLeft(&nb.sendQueue, nq)
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
// Close is thread-safe, non-blocking, error-free and panic-free.
func (nb *NBChan[T]) Close() (didClose bool) {
	if nb.isCloseInvoked.IsTrue() {
		return // Close was already invoked atomic performance
	}
	nb.stateLock.Lock()
	defer nb.stateLock.Unlock()

	if !nb.isCloseInvoked.Set() {
		return // Close was already invoked
	}

	if nb.isRunningThread {
		return // there is a pending thread that will execute close on exit
	}

	didClose, _ = nb.close()
	return
}

// DidClose indicates if Close was invoked
//   - the channel may remain open until the last item has been read
//   - or CloseNow is invoked
func (nb *NBChan[T]) DidClose() (didClose bool) {
	return nb.isCloseInvoked.IsTrue()
}

// IsClosed indicates whether the channel has actually closed.
func (nb *NBChan[T]) IsClosed() (isClosed bool) {
	return nb.closableChan.IsClosed()
}

// WaitForClose blocks until the channel is closed and empty
//   - if Close is not invoked or the channel is not read to end,
//     WaitForClose blocks indefinitely
//   - if CloseNow is invoked, WaitForClose is unblocked
//   - if errp is non-nil, any thread and close errors are appended to it
//   - a close error will already have been returned by Close
//   - WaitForClose is thread-safe and panic-free
func (nb *NBChan[T]) WaitForClose(errp ...*error) {
	defer nb.appendErrors(nil, nb.GetError, errp...)

	if nb.closableChan.IsClosed() {
		return // channel is closed no wait required return
	}
	if !nb.isWaitForCloseInitialized.IsTrue() {
		nb.initWaitForClose() // ensure waitForClose state is valid
	}
	nb.waitForClose.Wait()
}

// CloseNow closes without waiting for sends to complete.
//   - CloseNow is thread-safe, non-blocking, error-free and panic-free
//   - CloseNow does not return until the channel is closed.
//   - Upon return, all invocations have a possible close error in err.
//   - if errp is non-nil, it is updated with error status
func (nb *NBChan[T]) CloseNow(errp ...*error) (didClose bool, err error) {
	if nb.closableChan.IsClosed() {
		return // channel is already closed
	}
	nb.stateLock.Lock()
	defer nb.stateLock.Unlock()

	nb.isCloseInvoked.Set()

	// discard pending data
	var length = len(nb.sendQueue)
	nb.sendQueue = nil
	if length > 0 {
		nb.unsentCount -= length
	}

	// empty the channel
	if nb.isRunningThread {
		select {
		case <-nb.closesOnThreadSend:
		case <-nb.Ch():
		}
	}

	didClose, err = nb.close()

	// save close error in possible error pointer
	nb.appendErrors(err, nil, errp...)

	return
}

func (nb *NBChan[T]) appendErrors(err error, getError func() (err error), errp ...*error) {

	// obtain error pointer
	var errp0 *error
	if len(errp) > 0 {
		errp0 = errp[0]
	}
	if errp0 == nil {
		return // no error pointer nowhere to store return
	}

	perrors.AppendErrorDefer(errp0, &err, getError)
}

// startThread launches the send thread
//   - must be invoked inside the lock
func (nb *NBChan[T]) startThread(value T) {
	nb.isRunningThread = true
	nb.closesOnThreadSend = make(chan struct{})
	go nb.sendThread(value) // send err in new thread
}

// sendThread operates outside the lock for as long as there are items to send
func (nb *NBChan[T]) sendThread(value T) {
	defer Recover(Annotation(), nil, func(err error) {
		if pruntime.IsSendOnClosedChannel(err) {
			return // ignore if the channel was or became closed
		}
		nb.AddError(err)
	})

	var ok bool
	for {
		nb.threadSend(value)

		if value, ok = nb.valueToSend(); !ok {
			return
		}
	}
}

func (nb *NBChan[T]) threadSend(value T) {
	defer close(nb.closesOnThreadSend)

	nb.closableChan.Ch() <- value // may block or panic
}

func (nb *NBChan[T]) valueToSend() (value T, ok bool) {
	nb.stateLock.Lock()
	defer nb.stateLock.Unlock()

	// clear isRunningThread inside the lock
	// possibly invoke close
	defer nb.sendThreadDefer(&ok)

	// count the item just sent
	nb.unsentCount--

	// no more values: end thread
	if ok = len(nb.sendQueue) != 0; !ok {
		return // thread to exit: ok == false return
	}

	// send next value in queue
	value = nb.sendQueue[0]
	pslices.TrimLeft(&nb.sendQueue, 1)
	// closesOnThreadSend is re-created every time prior to exiting the lock
	// when a value is provided for the send thread
	nb.closesOnThreadSend = make(chan struct{})
	return
}

// sendThreadDefer is invoked when send thread is about to exit
//   - sendThreadDefer is invoked inside the lock
func (nb *NBChan[T]) sendThreadDefer(ok *bool) {
	if *ok {
		return // thread is not exiting
	}
	nb.isRunningThread = false
	if nb.isCloseInvoked.IsTrue() { // Close() was invoked after thread started
		nb.close()
	}
}

// close closes the underlying channel
//   - close is invoked inside the lock
func (nb *NBChan[T]) close() (didClose bool, err error) {
	if didClose, err = nb.closableChan.Close(); !didClose {
		return
	}
	if err != nil {
		nb.AddError(err) // store possible close error
	}
	if nb.isWaitForCloseInitialized.IsTrue() {
		nb.waitForClose.Done()
	}

	return
}

// initWaitForClose ensures that waitForClose is valid
func (nb *NBChan[T]) initWaitForClose() {
	nb.stateLock.Lock()
	defer nb.stateLock.Unlock()

	if nb.closableChan.IsClosed() {
		return // channel already closed
	}
	if !nb.isWaitForCloseInitialized.Set() {
		return // was already initialized return
	}

	nb.waitForClose.Add(1) // has to wait for close to occur
}
