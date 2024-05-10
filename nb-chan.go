/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
)

const (
	// T elements default allocation size for buffers unless specified
	defaultNBChanSize = 10
)

// NBChan is a non-blocking send, unbound-queue channel.
//   - NBChan behaves like a channel and a thread-safe slice
//   - — efficiency of sending and receiving multiple items at once
//   - — ability to wait for items to become available
//   - NBChan is initialization-free, thread-safe, idempotent and observable with panic-free and/or deferrable methods
//   - values are sent non-blocking, panic free and error free using:
//   - — [NBChan.Send] for single item
//   - — [NBChan.SendMany] for any number of items
//   - values are received from NBChan via:
//   - — a Go receive channel returned by [NBChan.Ch] or
//   - — fetched all, one or many at once using [NBChan.Get]
//   - — for Get, values can be awaited using [NBChan.DataWaitCh]
//   - [NBChanThreadType] provided to [NewNBChan] configures for performance:
//   - — [NBChanNone]: highest throughput at lowest cpu load
//   - — — cost is no channel sending values, ie. Ch is not available
//   - — — only way to receive items is [NBChan.Get]
//   - — — Get returns any number of items at once
//   - — — wait is by [NBChan.DataWaitCh]
//   - — — benefit is no thread
//   - — [NBChanAlways] is higher throughput than regular thread
//   - — — cost is thread is always running
//   - — — Ch Get and DataWaitCh are all available
//   - — with regular thread or [NBChanAlways]:
//   - — — [NBChan.Ch] offers wait with channel receive
//   - — — [NBChan.DataWaitCh] offers wait for data available
//   - low-allocation throughput can be obtained by using NBChan to handle
//     slices of items []Value. NBChan can then operate at near zero allocations.
//   - NBChan has deferrable, panic-free, observable, idempotent close.
//     The underlying channel is closed when:
//   - — [NBChan.Close] is invoked and the channel is read to empty, ie.deferred close
//   - — [NBChan.CloseNow] is invoked
//   - NBChan is observable:
//   - — [NBCHan.DidClose] indicates whether Close or CloseNow has been invoked
//   - — [NBChan.IsClosed] indicates whether the underlying channel has closed
//   - — [NBChan.WaitForClose] is deferrable and panic-free and waits until the underlying channel has been closed.
//   - — [NBChan.WaitForCloseCh] returns a channel that closes when the underlying channel closes
//   - NBChan is designed for error-free operation and only has panics and close errrors. All errors can be collected via:
//   - — [NBChan.CloseNow] [NBChan.WaitForClose] or [NBChan.GetError]
//   - NBChan has contention-separation between Send/SendMany and Get
//   - NBChan used as an error channel avoids the sending thread blocking
//     from a delayed or missing reader.
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
//	defer parl.Recover(parl."", &err, errCh.AddErrorProc)
//	errCh.Ch() <- err // non-blocking
//	if err = someFunc(); err != nil {
//	err = perrors.Errorf("someFunc: %w", err)
//	return
type NBChan[T any] struct {
	// data items can be retreived one at a time by receiving from this channel via [NBChan.Ch]
	//	- NBChan must be configured to have thread
	closableChan ClosableChan[T]
	// size to use for [NBChan.newQueue]
	allocationSize atomic.Uint64
	// number of items held by NBChan
	//	- one item may be with [NBChan.sendThread]
	//	- only incremented by Send SendMany when appending to input queue
	//	- decremented by sendThread when value sent on channel
	//	- decreased by Get when removing from output buffer
	//	- set to zero by CloseNow
	//	- may exit on-demand thread on reaching zero
	unsentCount atomic.Uint64
	// number of pending [NBChan.Get] invocations
	//	- blocks [NBChan.sendThread] from fetching more values
	gets atomic.Uint64
	// holds thread waiting while gets > 0
	//	- [NBChan.CloseNow] uses getsWait to await Get conclusion
	//	- executeChClose uses getsWait to await Get conclusion
	//	- thread uses getsWait to reduce outputLock contention by
	//		not retrieving values while Get invocations in progress or holding at lock
	getsWait PeriodWaiter
	// number of pending [NBChan.Send] [NBChan.SendMany] invocations
	sends atomic.Uint64
	// prevents thread from exiting while sends > 0
	//	- [NBChan.CloseNow] uses sendsWait to await Send SendMany conclusion
	//	- executeChClose uses sendsWait to await Send SendMany conclusion
	//	- thread uses sendsWait to await the conclusion of possible Send SendMany before
	//		checking for another item
	sendsWait PeriodWaiter
	// capacity of [NBChan.inputQueue]
	//	- written behind inputLock
	inputCapacity atomic.Uint64
	// capacity of [NBChan.outputQueue]
	//	- written behind outputLock
	outputCapacity atomic.Uint64
	// indicates threadless NBChan
	//	- from [NBChanNone] or [NBChan.SetNoThread]
	//	- [NBChan.Ch] is unavailable
	//	- [NBChan.DataWaitCh] is used for wait
	noThread atomic.Bool
	// a channel that closes when data is available
	dataWaitCh atomic.Pointer[chan struct{}]
	// makes data channel wait operations executing serially
	availableLock sync.Mutex
	// written behind availableLock
	isDataAvailable atomic.Bool
	// indicates thread always running, ie. no on-demand
	//	- from [NBChanAlways] or [NBChan.SetAlwaysThread]
	isThreadAlways atomic.Bool
	// wait mechanic for always-thread alert-wait
	threadCh       chan *T
	threadCh2      chan struct{}
	threadChWinner atomic.Bool
	// true if a thread was ever launched
	tcDidLaunchThread atomic.Bool
	// tcThreadLock atomizes tcRunningThread access with other actions:
	//	- tcStartThreadWinner: make set to true serialized execution
	//	- selectCloseNowWinner: atomize with isCloseNow and isCloseInvoked
	//	- selectCloseWinner: atomize with isCloseInvoked
	//	- collectSendThreadValue: determining thread running and collecting its value
	//	- isClose detection for alway-running alert-wait-entry
	//	- sendThreadExitCheck: make set to false serialized execution
	//	- sendThreadIsCloseNow: atomize set to false with isCloseNow
	tcThreadLock sync.Mutex
	// tcRunningThread indicates that [NBChan.sendThread] is running or launching
	//	- set to true when background decides to launch the thread
	//	- selects winner to invoke [NBChan.tcStartThread]
	//	- written behind threadLock
	//	- set to false by thread when:
	//	- — CloseNow detected
	//	- — on-demand thread or an always-thread detecting Close that encounters:
	//		unsent-count zero with no ongoing Send SendMany
	tcRunningThread atomic.Bool
	// tcThreadExitAwaitable makes the thread awaitable
	//	- armed behind tcThreadLock
	//	- used by CloseNow to ensure thread exit
	tcThreadExitAwaitable CyclicAwaitable
	tcSendBlock           atomic.Bool
	tcCollectRequest      atomic.Bool
	collectorLock         sync.Mutex
	collectLock           sync.Mutex
	collectChan           LacyChan[struct{}]
	// inputLock controls input: inputQueue and close
	//	 - makes mutually exlusive anything affecting the input buffer:
	//		[NBChan.Close] [NBChan.CloseNow] swapQueues state-access
	//		[NBChan.Scavenge] [NBChan.Send] [NBChan.SendMany]
	inputLock sync.Mutex
	// behind inputLock. One item may be with sendThread
	inputQueue []T
	//	A winner of Close or CloseNow was selected
	//	- close may be deferred while NBChan is not empty
	//	- written behind inputLock
	isCloseInvoked atomic.Bool
	// A winner of CloseNow was selected
	//	- written behind inputLock
	isCloseNow OnceCh
	// mechanic to wait for underlying channel close complete
	waitForClose Awaitable
	// getProgressLock ensures that if Send SendMany detects Get in progress,
	// on Get conclusion, the last Get is guaranteed to check threadProgressRequired
	getProgressLock sync.Mutex
	// getProgressLock bundles write to threadProgressRequired with
	// its justifying action
	progressLock sync.Mutex
	// set by Send and SendMany at any time if:
	//	- NBChan was empty
	//	- on-demand or always-on threading is used
	//	- a thread could not be started or notified
	threadProgressRequired atomic.Bool
	// tcState is a channel that sends state names when thread is in static hold
	//	- returned by [NBChan.StateCh]
	tcState atomic.Pointer[chan NBChanTState]
	// outputLock controls output functions
	//	- must not be acquired while holding inputLock
	//	- makes mutually exclusive anything affecting the output buffer:
	//		[NBChan.Get] [NBCHan.CloseNow] state-operation
	//		[NBChan.Scavenge] thread-send, output-allocation-to-size
	outputLock sync.Mutex
	// behind getLock: outputQueue initial state, sliced to zero length
	outputQueue0 []T
	// behind getLock: outputQueue sliced off from low to high indexes
	outputQueue []T
	// thread panics and channel close errors
	perrors.ParlError
}

// NewNBChan returns a non-blocking trillion-size buffer channel.
//   - NewNBChan allows initialization based on an existing channel.
//   - NBChan does not need initialization and can be used like:
//
// Usage:
//
//	var nbChan NBChan[error]
//	go thread(&nbChan)
func NewNBChan[T any](threadType ...NBChanThreadType) (nbChan *NBChan[T]) {
	n := NBChan[T]{}
	if len(threadType) > 0 {
		switch threadType[0] {
		case NBChanAlways:
			n.isThreadAlways.Store(true)
		case NBChanNone:
			n.noThread.Store(true)
		}
	}
	return &n
}

// SetAllocationSize sets the initial element size of the two queues. Thread-safe
//   - NBChan allocates two queues of size which may be enlarged by item counts
//   - supports functional chaining
//   - 0 or less does nothing
func (n *NBChan[T]) SetAllocationSize(size int) (nb *NBChan[T]) {
	nb = n
	if size <= 0 {
		return // noop return
	}
	n.allocationSize.Store(uint64(size))
	n.ensureInput(size)
	n.ensureOutput(size)
	return
}

// SetAlwaysThread configures [NBChanAlways] operation
func (n *NBChan[T]) SetAlwaysThread() (nb *NBChan[T]) {
	nb = n
	n.isThreadAlways.Store(true)
	return
}

// SetNoThread configures [NBChanNone] operation
func (n *NBChan[T]) SetNoThread() { n.noThread.Store(true) }

const (
	// [NBChan.ThreadStatus] await a blocked thread state or exit
	AwaitThread = true
)

// ThreadStatus indicates the current status of a possible thread
func (n *NBChan[T]) ThreadStatus(await ...bool) (threadStatus NBChanTState) {
	if len(await) > 0 && await[0] {
		// await thread status
		//	- thread not launched
		if !n.tcDidLaunchThread.Load() {
			threadStatus = NBChanNoLaunch
			return
		}
		select {
		// status from a blocked thread
		//	- NBChanSendBlock NBChanAlert NBChanGets NBChanSends
		case threadStatus = <-n.stateCh():
			// thread did exit
		case <-n.tcThreadExitAwaitable.Ch():
			threadStatus = NBChanExit
		}
		return
	}

	// obtain current thread status, including running: ie. no idea where
	select {
	// status from a blocked thread
	//	- NBChanSendBlock NBChanAlert NBChanGets NBChanSends
	case threadStatus = <-n.stateCh():
		// thread did exit
	case <-n.tcThreadExitAwaitable.Ch():
		threadStatus = NBChanExit
	default:
		// thread is somewhere else
		if n.tcDidLaunchThread.Load() {
			threadStatus = NBChanRunning
		} else {
			threadStatus = NBChanNoLaunch
		}
	}
	return
}

// Ch obtains the receive-only channel
//   - values can be retrieved using this channel or [NBChan.Get]
//   - not available for [NBChanNone] NBChan
func (n *NBChan[T]) Ch() (ch <-chan T) { return n.closableChan.Ch() }

// Count returns number of unsent values
func (n *NBChan[T]) Count() (unsentCount int) { return int(n.unsentCount.Load()) }

// Capacity returns size of allocated queues
func (n *NBChan[T]) Capacity() (capacity int) {
	return int(n.inputCapacity.Load() + n.outputCapacity.Load())
}

// DataWaitCh is a method to wait for additional values
//   - alternative is [NBChan.Ch]
//   - DataWaitCh offers more efficient operation
func (n *NBChan[T]) DataWaitCh() (ch AwaitableCh) { return n.updateDataAvailable() }

// DidClose indicates if Close or CloseNow was invoked
//   - the channel may remain open until the last item has been read
//   - [NBChan.CloseNow] immediately closes the channel discarding onread items
//   - [NBChan.IsClosed] checks if the channel is closed
func (n *NBChan[T]) DidClose() (didClose bool) { return n.isCloseInvoked.Load() }

// IsClosed indicates whether the channel has actually closed.
func (n *NBChan[T]) IsClosed() (isClosed bool) { return n.closableChan.IsClosed() }
