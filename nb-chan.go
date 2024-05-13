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
//   - see also:
//   - — [AwaitableSlice] unbound awaitable queue
//   - — [NBRareChan] low-usage unbound channel
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
	// [NBChan.Ch] returns this channel allowing consumers to await data items one at a time
	//	- NBChan must be configured to have thread
	closableChan ClosableChan[T]
	// size to use for [NBChan.newQueue]
	allocationSize atomic.Uint64
	// number of items held by NBChan, updated at any time
	//	- [NBChan.sendThread] may hold one item
	//	- only incremented by Send SendMany while holding
	//		inputLock appending to input queue or handing value to thread.
	//		Increment may be delegated to always-thread
	//	- decreased by Get when removing from output buffer while
	//		holding outputLock
	//	- decremented by send-thread when value sent on channel
	//	- decremented by send-thread when detecting CloseNow
	//	- set to zero by CloseNow while holding outputLock
	//	- a on-demand thread or deferred-close always-thread may
	//		exit on observing unsent count zero
	unsentCount atomic.Uint64
	// number of pending [NBChan.Get] invocations
	//	- blocks [NBChan.sendThread] from fetching more values
	//	- awaitable via getsWait
	gets atomic.Uint64
	// holds thread waiting while gets > 0
	//	- [NBChan.CloseNow] uses getsWait to await Get conclusion
	//	- executeChClose uses getsWait to await Get conclusion
	//	- thread uses getsWait to reduce outputLock contention by
	//		not retrieving values while Get invocations in progress or holding at lock
	getsWait PeriodWaiter
	// number of pending [NBChan.Send] [NBChan.SendMany] invocations
	//	- awaitable via sendsWait
	sends atomic.Uint64
	// prevents thread from exiting while Send SendMany active
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
	isNoThread atomic.Bool
	// a channel that closes when data is available
	dataWaitCh atomic.Pointer[chan struct{}]
	// makes data channel wait operations executing serially
	availableLock sync.Mutex
	// written behind availableLock
	isDataAvailable atomic.Bool
	// indicates thread always running, ie. no on-demand
	//	- from [NBChanAlways] or [NBChan.SetAlwaysThread]
	isOnDemandThread atomic.Bool
	// wait mechanic with value for always-thread alert-wait
	//	- used in two-chan send with threadCh2
	alertChan LacyChan[*T]
	// second channel for two-chan send with threadCh
	alertChan2       LacyChan[struct{}]
	alertChan2Active atomic.Pointer[chan struct{}]
	// tcAlertActive ensures one alert action per alert wait
	//	- sendThreadWaitForAlert: sets to true while awaiting
	//	- tcAlertThread: picks winner for alerting
	//	- two-chan send with threadCh threadCh2
	tcAlertActive atomic.Bool
	// true if a thread was ever launched
	//	- [NBChan.ThreadStatus] uses tcDidLaunchThread to distinguish between
	//		NBChanNoLaunch and NBChanRunning
	tcDidLaunchThread atomic.Bool
	// tcThreadLock atomizes tcRunningThread with other actions:
	//	- tcThreadLock enforces order so that no thread is created after CloseNow
	//	- tcStartThreadWinner: atomize isCloseNow detection with setting tcRunningThread to true
	//	- selectCloseNowWinner: atomize tcRunningThread read with isCloseNow and isCloseInvoked set to true
	//	- selectCloseWinner: atomize tcRunningThread read with setting isCloseInvoked to true
	//	- sendThreadExitCheck: make seting tcRunningThread to false mutually exclusive with other operations
	//	- sendThreadIsCloseNow: atomize isCloseNow detection with setting tcRunningThread to false
	tcThreadLock sync.Mutex
	// tcRunningThread indicates that [NBChan.sendThread] is about to be created or running
	//	- set to true when background decides to launch the thread
	//	- set to false when:
	//	- — send-thread detects CloseNow
	//	- — an on-demand thread or a deferred-close always-thread encountering:
	//	- — unsent-count zero with no ongoing Send SendMany
	//	- selects winner to invoke [NBChan.tcStartThread]
	//	- written behind tcThreadLock
	tcRunningThread atomic.Bool
	// tcThreadExitAwaitable makes the thread awaitable
	//	- tcStartThreadWinner: arms or re-arms behind tcThreadLock
	//	- triggered by thread on exit
	//	- [NBChan.CloseNow] awaits tcThreadExitAwaitable to ensure thread exit
	//	- used in two-chan receive with tcState for awaiting static thread-state
	tcThreadExitAwaitable CyclicAwaitable
	// tcSendBlock true indicates that send-thread has value
	// availble for two-chan receive from underlying channel and collectChan
	tcSendBlock atomic.Bool
	// collectorLock ensures any alert-thread will not be alerted while
	// Get is in progress
	//	- preGet: on transition 0 to 1 Get, acquires collectorLock to establish Get in progress
	//	- tcAlertThread: atomizes observing gets zero with alert operation
	collectorLock sync.Mutex
	// collectChan is used in two-chan receive with underlying channel
	// by tcCollectThreadValue
	collectChan LacyChan[struct{}]
	// collectChanActive is the channel being used by two-chan receive
	// from underlying channel
	collectChanActive atomic.Pointer[chan struct{}]
	// inputLock makes inputQueue thread-safe
	//	- used by: [NBChan.CloseNow] swapQueues [NBChan.NBChanState]
	//		[NBChan.Scavenge] [NBChan.Send] [NBChan.SendMany]
	//		[NBChan.SetAllocationSize]
	//	- isCloseComplete acquires inputLock to Send SendMany
	//		have ceased and no further will invocations commence
	inputLock sync.Mutex
	// inputQueue holds items from Send SendMany
	//	- access behind inputLock
	//	- one additional item may be with sendThread
	inputQueue []T
	// isCloseInvoked selects Close winner thread and provides Close invocation wait
	//	- isCloseInvoked.IsWinner select winner
	//A winner of Close or CloseNow was selected
	//	- for Close, close may be deferred while NBChan is not empty
	//	- written behind tcThreadLock to ensure no further thread launches
	isCloseInvoked OnceCh
	// A winner of CloseNow was selected
	//	- written inside threadLock
	isCloseNow OnceCh
	// mechanic to wait for underlying channel close complete
	//	- the underlying channel is closed by:
	//	- — non-deferred Close
	//	- — send-thread in deferred Close
	//	- CloseNow
	waitForClose Awaitable
	// tcProgressLock atomizes tcProgressRequired updates with their justifying observation
	tcProgressLock sync.Mutex
	// tcProgressRequired indicates that thread progress must be secured
	//	- tcAddProgress: set to true by Send SendMany when adding from unsent count zero
	//	- sendThreadZero: set to true by send-thread when taking action on unsent count zero
	//	- tcLaunchProgress: set to false on obtaining thread launch permisssion
	//	- tcAlertProgress: set to false on successful alert
	//	- tcThreadStateProgress: set to false on observing any thread state but NBChanAlert
	//	- write behind tcProgressLock
	tcProgressRequired atomic.Bool
	// tcProgressRaised notes intermediate events:
	//	- send-thread taking action on unsent count zero
	//	- Send SendMany increasing unsent count from zero
	tcProgressRaised atomic.Bool
	// tcAwaitProgressLock makes tcAwaitProgress a critical section
	//	- only one thread at a time may await the next static thread state
	tcAwaitProgressLock sync.Mutex
	// tcGetProgressLock atomizes read of tcProgressRequired and pending Get
	//	- this ensures that when progress is required while Get in progress,
	//		action is guaranteed by Get on the last invocation ending
	tcGetProgressLock sync.Mutex
	// tcState is a channel that sends state names when thread is in static hold
	//	- returned by [NBChan.StateCh]
	tcState atomic.Pointer[chan NBChanTState]
	// outputLock makes outputQueue thread-safe
	//	- must not be acquired while holding inputLock
	//	- used by: [NBChan.Get] [NBChan.CloseNow] [NBChan.NBChanState]
	//		[NBChan.Scavenge] [NBChan.SetAllocationSize]
	//	- used by thread to ontain next value when Get not in progress
	outputLock sync.Mutex
	// behind outputLock: outputQueue initial state, sliced to zero length
	outputQueue0 []T
	// behind outputLock: outputQueue sliced off from low to high indexes
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
		case NBChanOnDemand:
			n.isOnDemandThread.Store(true)
		case NBChanNone:
			n.isNoThread.Store(true)
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

// SetOnDemandThread configures [NBChanAlways] operation
func (n *NBChan[T]) SetOnDemandThread() (nb *NBChan[T]) {
	nb = n
	n.isOnDemandThread.Store(true)
	n.isNoThread.Store(false)
	return
}

// SetNoThread configures [NBChanNone] operation
func (n *NBChan[T]) SetNoThread() {
	n.isNoThread.Store(true)
	n.isOnDemandThread.Store(false)
}

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

// DataWaitCh indicates if the NBChan object has data available
//   - the initial state of the returned channel is open, ie. receive will block
//   - upon data available the channel closes, ie. receive will not block
//   - to again wait, DataWaitCh should be invoked again to get a current channel
//   - upon CloseNow or Close and empty, the returned channel is closed
//     and receive will not block
//   - alternative is [NBChan.Ch] when NBChan configured with threading
//   - DataWaitCh offers more efficient operation that with a thread
func (n *NBChan[T]) DataWaitCh() (ch AwaitableCh) { return n.updateDataAvailable() }

// DidClose indicates if Close or CloseNow was invoked
//   - the channel may remain open until the last item has been read
//   - [NBChan.CloseNow] immediately closes the channel discarding onread items
//   - [NBChan.IsClosed] checks if the channel is closed
func (n *NBChan[T]) DidClose() (didClose bool) { return n.isCloseInvoked.IsInvoked() }

// IsClosed indicates whether the channel has actually closed.
func (n *NBChan[T]) IsClosed() (isClosed bool) { return n.closableChan.IsClosed() }
