/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/sets"
)

const (
	defaultNBChanSize = 10         // T elements
	NBChanExit        = "exit"     // NBChan thread is not running
	NBChanAlert       = "alert"    // NBChan thread is always running and blocked idle waiting for alert
	NBChanGets        = "GetWait"  // NBChan thread is blocked waiting for Get invocations to complete
	NBChanSends       = "SendWait" // NBChan thread is blocked waiting for Send/SendMany invocations to complete
	NBChanSendBlock   = "chSend"   // NBChan thread is blocked in channel send
	NBChanRunning     = "run"      // NBChan is running
)

const (
	// NBChanAlways configures NBChan to always have a thread
	//   - benefit: for empty NBChan, Send/SendMany to  channel receive is faster due to avoiding thread launch
	//   - cost: a thread is always running instad of only running when NBChan non-empty
	//   - cost: Close or CloseNow must be invoked to shutdown NBChan
	NBChanAlways NBChanThreadType = iota + 1
	// NBChanNone configures no thread.
	//	- benefit: lower cpu
	//	- cost: data can only be received using [NBChan.Get]
	NBChanNone
)

// NBChan is a non-blocking send channel with trillion-size queues.
//   - NBChan behaves both like a channel and a thread-safe slice
//   - — efficiency of sending and receiving multiple items at once
//   - — ability to wait for items to become available
//   - NBChan is initialization-free, thread-safe, idempotent and observable with panic-free methods and deferrable medthods
//   - values are sent to the channel using Send/SendMany that are never blocked by channel send
//     and are panic-free error-free
//   - values are received from the channel or fetched all or many at once using Get
//   - — Get wait:DataWaitCh and NBChanNone is highest throughput at lowest cpu load
//   - — — cost is no channel is sending values, Get must be used
//   - — — benefit is no thread
//   - — Get Ch and NBChanAlways is higher throughput than regular thread
//   - — — cost is thread is always running
//   - — with regular thread or NBChanAlways:
//   - — — Ch offers wait with channel receive
//   - — — DataWaitCh only waits for data available
//   - NBChan has deferrable, panic-free, observable, idempotent close.
//     The underlying channel is closed when:
//   - — Close is invoked and the channel is read to end
//   - — CloseNow is invoked
//   - NBChan is observable:
//   - — DidClose indicates whether Close or CloseNow has been invoked
//   - — IsClosed indicates whether the underlying channel has closed
//   - — WaitForClose is deferrable and panic-free and waits until the underlying channel has been closed.
//   - — WaitForCloseCh returns a channel that closes when the underlying channel closes
//   - NBChan is designed for error-free operation and only has panics and close errrors. All errors can be collected via:
//   - — CloseNow WaitForClose GetError
//   - NBChan has contention-separation between Send/SendMany and Get
//   - NBChan can be used as an error channel where the sending thread does not
//     block from a delayed or missing reader.
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

	// size to use for [NBChan.newQueue]
	//	- atomic
	//	- set by [NBChan.SetAllocationSize]
	allocationSize uint64
	// number of items held by NBChan
	//	- one item may be with [NBChan.sendThread]
	//	- atomic
	unsentCount uint64
	// number of pending [NBChan.Get] invocations
	//	- blocks [NBChan.sendThread] from fetching more values
	//	- atomic
	gets      uint64
	getsWait  PeriodWaiter // holds thread waiting while gets > 0
	sends     uint64
	sendsWait PeriodWaiter // prevents thread from exiting while sends > 0
	// capacity of [NBChan.inputQueue]
	//	- atomic
	//	- written behind inputLock
	inputCapacity uint64
	// capacity of [NBChan.outputQueue]
	//	- atomic
	//	- written behind outputLock
	outputCapacity uint64
	noThread       atomic.Bool
	dataWaitCh     atomic.Pointer[chan struct{}]

	availableLock   sync.Mutex
	isDataAvailable atomic.Bool // written behind availableLock

	// allows to wait for thread exit
	//	- because thread may relaunch only relevent after isCloseInvoked
	threadWait atomic.Pointer[chan struct{}]
	// closesOnThreadSend is a channel created prior to every
	// sendThread channel send operation
	//	- when isRunningThread true, closesOnThreadSend is valid
	//	- the channel will close immediately after thread send completes
	//	- this allows for the thread’s value to be collected by Get
	//	- reinitialized behind threadLock
	closesOnThreadSend atomic.Pointer[chan struct{}]
	isThreadAlways     atomic.Bool
	// set to true by sendThread when entering alert-wait
	//	- reset by winner invocation for alerting sendThread
	threadAlertPending atomic.Bool
	threadAlertWait    PeriodWaiter      // wait mechanic for always-thread alert-wait
	threadAlertValue   atomic.Pointer[T] // possible value sent by winner invocation to sendThread
	// threadLock bundles isRunningThread updates:
	//	- isRunningThread false-to-true winner and closesOnThreadSend reinitialized
	//	- reinitialize closesOnThreadSend
	//	- isClose isCloseNow transitions and isRunningThread state
	//	- collecting sendThread value with isRunningThread and closesOnThreadSend
	//	- isClose detection for alway-running alert-wait-entry
	//	- sendThread exit decision
	//	- sendThread isCloseNow detection
	threadLock sync.Mutex
	// isRunningThread indicates that [NBChan.sendThread] is running
	//	- winner gets to invoke [NBChan.startThread]
	isRunningThread atomic.Bool // written behind threadLock

	inputLock         sync.Mutex  // controls inputQueue and close
	inputQueue        []T         // behind inputLock. One item may be with sendThread
	isCloseInvoked    atomic.Bool // written behind inputLock
	isCloseNowInvoked atomic.Bool // written behind inputLock

	isWaitForCloseDone atomic.Bool                   // decides winner to close waitForClose
	waitForClose       atomic.Pointer[chan struct{}] // mechanic to wait for close completion

	outputLock   sync.Mutex // must not be acquired while holding inputLock
	outputQueue0 []T        // behind getLock: outputQueue initial state, sliced to zero length
	outputQueue  []T        // behind getLock: outputQueue sliced off from low to high indexes

	perrors.ParlError // thread panics and channel close errors
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
func (n *NBChan[T]) SetAllocationSize(size int) (nb *NBChan[T]) {
	nb = n
	if size <= 0 {
		return // noop return
	}
	atomic.StoreUint64(&n.allocationSize, uint64(size))
	n.ensureInput(size)
	n.ensureOutput(size)
	return
}

func (n *NBChan[T]) SetAlwaysThread() (nb *NBChan[T]) {
	nb = n
	n.isThreadAlways.Store(true)
	return
}

func (n *NBChan[T]) SetNoThread() {
	n.noThread.Store(true)
}

// ThreadStatus indicates the current status of a possible thread
func (n *NBChan[T]) ThreadStatus() (threadStatus string) {
	if !n.isRunningThread.Load() {
		return NBChanExit
	} else if n.threadAlertPending.Load() {
		return NBChanAlert
	} else if n.getsWait.Count() > 0 {
		return NBChanGets
	} else if n.sendsWait.Count() > 0 {
		return NBChanSends
	} else if n.isSendThreadChannelSend() {
		return NBChanSendBlock
	}
	return NBChanRunning
}

// Ch obtains the receive-only channel
//   - values can be retrieved using this channel or [NBChan.Get]
func (n *NBChan[T]) Ch() (ch <-chan T) {
	return n.closableChan.Ch()
}

// Count returns number of unsent values
func (n *NBChan[T]) Count() (unsentCount int) {
	return int(atomic.LoadUint64(&n.unsentCount))
}

// Capacity returns size of allocated queues
func (n *NBChan[T]) Capacity() (capacity int) {
	return int(atomic.LoadUint64(&n.inputCapacity) + atomic.LoadUint64(&n.outputCapacity))
}

func (n *NBChan[T]) DataWaitCh() (ch <-chan struct{}) {
	return n.updateDataAvailable()
}

// DidClose indicates if Close or CloseNow was invoked
//   - the channel may remain open until the last item has been read
//   - [NBChan.CloseNow] immediately closes the channel discarding onread items
//   - [NBChan.IsClosed] checks if the channel is closed
func (n *NBChan[T]) DidClose() (didClose bool) {
	return n.isCloseInvoked.Load()
}

// IsClosed indicates whether the channel has actually closed.
func (n *NBChan[T]) IsClosed() (isClosed bool) {
	return n.closableChan.IsClosed()
}

type NBChanThreadType uint8

func (n NBChanThreadType) String() (s string) {
	return nbChanThreadTypeSet.StringT(n)
}

var nbChanThreadTypeSet = sets.NewSet(sets.NewElements[NBChanThreadType](
	[]sets.SetElement[NBChanThreadType]{
		{ValueV: NBChanAlways, Name: "alwaysCh"},
		{ValueV: NBChanNone, Name: "noCh"},
	}))
