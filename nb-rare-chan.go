/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
)

// NBRareChan is a simplified [NBChan] using on-demand thread
//   - NBRareChan is a channel with unbound buffer
//     like an unbuffered channel reading from a thread-safe slice
//   - [NBRareChan.Send] provides value send that is non-blocking-send, thread-safe, panic, deadlock and error-free
//   - [NBRareChan.Ch] provides real-time value stream or
//   - [NBRareChan.Close] provides all buffered values
//   - [NBRareChan.StopSend] blocks further Send allowing for graceful shutdown
//   - [NBRareChan.IsClose] returns whether the underlying channel is closed
//   - [NBRareChan.PanicCh] provides real-time notice of thread panic, should not happen
//   - NBRareChan is initialization-free and thread-safe with
//     thread-safe panic-free idempotent observable deferrable Close
//   - used as an error sink, NBRareChan[error] prevents error propagation from affecting the thread
//   - ignoring thread panics and Close errp is reasonably safe simplification
//   - intended for infrequent use such as an error sink
//   - benefits over plain channel:
//   - — [NBRareChan.Send] is non-blocking-send panic-free non-dead-locking
//   - — initialization-free
//   - — unbound buffer
//   - — any thread can close the channel as opposed to only the sending thread
//   - — thread-synchronizing unbuffered channel-send mechanic as opposed to a buffered channel
//   - — graceful shutdown like a buffered channel
//   - — [NBRareChan.Close] is any-thread data-race-free thread-safe panic-free idempotent observable deferrable
//   - drawbacks compared to [NBChan]:
//   - — on-demand thread may lead to high cpu if used frequently like every second.
//     NBChan offers no-thread or always-thread operation
//   - — buffering a large number of items leads to a temporary memory leak in queue
//   - — there is no contention-separation between Send and reading Ch
//   - — no multiple-item operations like SendMany or Get
//   - — less observable and configurable
//   - see also:
//   - — [NBChan] fully-featured unbound channel
//   - — [AwaitableSlice] unbound awaitable queue
//
// Deprecated: NBRareChan is replaced by [github.com/haraldrudell/parl.AwaitableSlice] for performance and
// efficiency reasons. [github.com/haraldrudell/parl.ErrSlice] is an error container implementation
type NBRareChan[T any] struct {
	// underlying channel
	//	- closed by first Close invocation
	closableChan ClosableChan[T]
	// threadWait makes all created send threads awaitable
	threadWait sync.WaitGroup
	// queueLock ensures thread-safety of queue
	//	- also ensures sequenced access to isThread isStopSend
	queueLock sync.Mutex
	// didSend idicates that Send did create a send-thread
	// whose value may need to be collected on Close
	//	- behind queueLock
	didSend bool
	// queue is a slice-away slice of unsent data
	//	- Send appends to queue
	//	- behind queueLock
	queue []T
	// threadReadingValues indicates that a send thread is
	// currently reading values from queue
	//	- accessed behind queueLock
	threadReadingValues CyclicAwaitable
	// accessed behind queueLock
	isStopSend atomic.Bool
	// returned by StopSend await empty channel
	isEmpty Awaitable
	// sendThread panics, should be none
	errs AtomicError
	// isPanic indicates that a send thread suffered a panic
	//	- triggers PanicCh awaitable
	isPanic Awaitable
	// ensures close executed once
	closeOnce OnceCh
	sink      nbrSink[T]
}

// Ch obtains the underlying channel for channel receive operations
func (n *NBRareChan[T]) Ch() (ch <-chan T) { return n.closableChan.Ch() }

// Send sends a single value on the channel
//   - non-blocking-send, thread-safe, deadlock-free, panic-free and error-free
//   - if Close or StopSend was invoked, value is discarded
func (n *NBRareChan[T]) Send(value T) {
	n.queueLock.Lock()
	defer n.queueLock.Unlock()

	// ignore values after Close or StopSend
	if n.closableChan.IsClosed() || n.isStopSend.Load() {
		return
	}

	// possibly create thread with value
	var createThread bool
	if createThread = !n.didSend; createThread {
		n.didSend = true
	} else {
		createThread = n.threadReadingValues.IsClosed()
	}
	if createThread {
		n.threadReadingValues.Open()
		n.threadWait.Add(1)
		go n.sendThread(value)
		return
	}

	// append value to buffer
	n.queue = append(n.queue, value)
}

// StopSend ignores further Send allowing for the channel to be drained
//   - emptyAwaitable triggers once the channel is empty
func (n *NBRareChan[T]) StopSend() (emptyAwaitable AwaitableCh) {
	n.queueLock.Lock()
	defer n.queueLock.Unlock()

	n.isStopSend.CompareAndSwap(false, true)
	emptyAwaitable = n.isEmpty.Ch()
	if len(n.queue) == 0 && n.threadReadingValues.IsClosed() {
		n.isEmpty.Close()
	}
	return
}

// PanicCh is real-time awaitable for panic in sendThread
//   - this should not happen
func (n *NBRareChan[T]) PanicCh() (emptyAwaitable AwaitableCh) { return n.isPanic.Ch() }

// IsClose returns true if underlying channel is closed
func (n *NBRareChan[T]) IsClose() (isClose bool) { return n.closableChan.IsClosed() }

// Close immediately closes the channel returning any unsent values
//   - values: possible values that were in channel, may be nil
//   - errp: receives any panics from thread. Should be none. may be nil
//   - upon return, resources are released and further Send ineffective
func (n *NBRareChan[T]) Close(values *[]T, errp *error) {

	// ensure once execution
	if isWinner, done := n.closeOnce.IsWinner(); !isWinner {
		return // loser thread has already awaited done
	} else {
		defer done.Done()
	}

	// collect queue and stop further Send
	var queue = n.close()

	// collect possible value from thread and shut it down
	if n.didSend {
		select {
		case value := <-n.closableChan.Ch():
			queue = append([]T{value}, queue...)
		case <-n.threadReadingValues.Ch():
			n.isEmpty.Close()
		}
		// wait for all created threads to exit
		n.threadWait.Wait()
		if errp != nil {
			if err, hasValue := n.errs.Error(); hasValue {
				*errp = perrors.AppendError(*errp, err)
			}
		}
	}

	// close underlying channel
	n.closableChan.Close()

	// return values
	if values != nil && len(queue) > 0 {
		*values = queue
	}
}

// collect queue and stop further Send
func (n *NBRareChan[T]) close() (values []T) {
	n.queueLock.Lock()
	defer n.queueLock.Unlock()

	if values = n.queue; values != nil {
		n.queue = nil
	}
	n.isStopSend.Store(true)

	return
}

// sendThread carries out send operations on the channel
func (n *NBRareChan[T]) sendThread(value T) {
	defer n.threadWait.Done()
	n.sink.r = n
	defer Recover(func() DA { return A() }, nil, &n.sink)

	var ch = n.closableChan.Ch()
	for {
		ch <- value
		var hasValue bool
		if value, hasValue = n.sendThreadNextValue(); !hasValue {
			return
		}
	}
}

// sendThreadNextValue obtains the next value to send for thread if any
func (n *NBRareChan[T]) sendThreadNextValue() (value T, hasValue bool) {
	n.queueLock.Lock()
	defer n.queueLock.Unlock()

	if hasValue = len(n.queue) > 0; hasValue {
		value = n.queue[0]
		n.queue = n.queue[1:]
		return
	}
	// channel detected empty

	//	- notify that sendThread is no longer awaiting values
	n.threadReadingValues.Close()

	// if StopSend was invoked, notify its awaitable
	if n.isStopSend.Load() {
		n.isEmpty.Close()
	}

	return
}

// sendThreadPanic aggregates thread panics
func (n *NBRareChan[T]) sendThreadPanic(err error) {
	n.errs.AddError(err)

	// notify awaitable that a panic occured
	n.isPanic.Close()
	n.queueLock.Lock()
	defer n.queueLock.Unlock()

	n.threadReadingValues.Close()
}

type nbrSink[T any] struct{ r *NBRareChan[T] }

func (n *nbrSink[T]) AddError(err error) { n.r.sendThreadPanic(err) }
