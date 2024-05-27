/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"
	"time"

	"github.com/haraldrudell/parl/ptime"
)

const (
	// disables the debounce time
	//	- debounce time holds incoming items until
	//		debounce time elapses with no additional items
	//	- when disabled max delay defaults to1 s and
	//		items are sent when maxDelay reached
	NoDebounceTime time.Duration = 0
	// disables debouncer max delay function
	//	- when debounce timer holds items, those items
	//		are sent when age reaches maxDelay
	//	- when debounce time disabled, defaults to 1 s.
	//		otherwise no default
	NoDebounceMaxDelay time.Duration = 0
	// maxDelay when debounce-time disabled
	defaultDebouncerMaxDelay = time.Second
)

// Debouncer debounces event stream values
//   - T values are received from the in channel
//   - Once d time has elapsed with no further incoming Ts,
//     a slice of read T values are provided to the sender function
//   - errFn receives any panics in the threads, expected none
//   - sender and errFn functions must be thread-safe.
//   - Debouncer is shutdown gracefully by closing the input channel or
//     immediately by invoking the Shutdown method
//   - —
//   - two threads are launched per debouncer
type Debouncer[T any] struct {
	in         *debouncerIn[T]  // input thread
	out        *debouncerOut[T] // output thread
	isShutdown *Awaitable       // shutdown control
}

// debouncerIn implements the debouncer input-thread
type debouncerIn[T any] struct {
	// from where incoming values for debouncing are read
	inputCh <-chan T
	// non-blocking unbound buffer to output thread
	buffer AwaitableSlice[T]
	// how long time must pass between two consecutive
	// incoming values in order to submit to output channel
	debounceInterval time.Duration
	// how input-thread orders output thread to send
	// on expired debounce period
	debounceTimer *time.Timer
	// is maxDelay timer is used
	useMaxDelay bool
	// when input thread receives a a value and max delay timer is not running,
	// max delay timer is started.
	//	- max delay timer then runs until output thread resets it.
	maxDelayRunning atomic.Bool
	// how input-thread orders output thread to send
	// on expired maxDelay period
	maxDelayTimer ptime.ThreadSafeTimer
	// how input thread receives shutdown
	isShutdown *Awaitable
	// how input thread emits an unforeseen panic
	errorSink ErrorSink1
	// awaitable indicating input thread exit
	inputExit Awaitable
}

// debouncerIn implements the debouncer input-thread
type debouncerOut[T any] struct {
	// non-blocking unbound buffer from input thread
	buffer *AwaitableSlice[T]
	// send trigger based on debounce time expired
	debounceC <-chan time.Time
	// is maxDelay timer is used
	useMaxDelay bool
	// when input thread receives a a value and max delay timer is not running,
	// max delay timer is started.
	//	- max delay timer then runs until output thread resets it.
	maxDelayRunning *atomic.Bool
	// maxDelayTimer timer expiring when output thread should send
	maxDelayTimer *ptime.ThreadSafeTimer
	// indicates that input thread exited
	isInputExit AwaitableCh
	// the output function receiving slices of values
	sender func([]T)
	// how output thread receives shutdown
	isShutdown *Awaitable
	// how output thread emits an unforeseen panic
	errorSink ErrorSink1
	// awaitable indicating output thread exit
	outputExit Awaitable
}

// NewDebouncer returns a channel debouncer
//   - values incoming faster than debounceInterval are aggregated
//     into slices
//   - values are not kept waiting longer than maxDelay
//   - debounceInterval is only used if > 0 ns
//   - if debounceInterval is not used and maxDelay is 0,
//     maxDelay defaults to 1 s to avoid a hanging debouncer
//   - sender should not be long-running or blocking
//   - inputCh sender errFn cannot be nil
//   - close of input channel or Shutdown is required to release resources
//   - errFn should not receive any errors but will receive possible runtime panics
//   - —
//   - NewDebouncer launches two threads prior to return
func NewDebouncer[T any](
	debounceInterval, maxDelay time.Duration,
	inputCh <-chan T,
	sender func([]T),
	errorSink ErrorSink1,
) (debouncer *Debouncer[T]) {
	if inputCh == nil {
		panic(NilError("inputCh"))
	} else if sender == nil {
		panic(NilError("sender"))
	} else if errorSink == nil {
		panic(NilError("errFn"))
	}

	var isShutdown Awaitable

	// debounce timer expiring when output thread should send
	var debounceTimer = time.NewTimer(time.Second)
	// get timer ready for reset
	debounceTimer.Stop()
	if len(debounceTimer.C) > 0 {
		<-debounceTimer.C
	}

	// 1 s default for maxDelay
	if debounceInterval <= 0 && maxDelay <= 0 {
		maxDelay = defaultDebouncerMaxDelay
	}

	in := debouncerIn[T]{
		inputCh:          inputCh,
		debounceInterval: debounceInterval,
		debounceTimer:    debounceTimer,
		useMaxDelay:      maxDelay > 0,
		maxDelayTimer:    *ptime.NewThreadSafeTimer(maxDelay),
		isShutdown:       &isShutdown,
		errorSink:        errorSink,
	}
	// get timer ready for reset
	in.maxDelayTimer.Stop()
	if len(in.maxDelayTimer.C) > 0 {
		<-in.maxDelayTimer.C
	}
	out := debouncerOut[T]{
		buffer:          &in.buffer,
		debounceC:       debounceTimer.C,
		useMaxDelay:     in.useMaxDelay,
		maxDelayRunning: &in.maxDelayRunning,
		maxDelayTimer:   &in.maxDelayTimer,
		isInputExit:     in.inputExit.Ch(),
		sender:          sender,
		isShutdown:      &isShutdown,
		errorSink:       errorSink,
	}

	go out.outputThread()
	go in.inputThread()

	return &Debouncer[T]{
		in:         &in,
		out:        &out,
		isShutdown: &isShutdown,
	}
}

// Shutdown shuts down the debouncer
//   - Shutdown does not return until resources have been released
//   - buffered values are discarded and input channle is not read to end
func (d *Debouncer[T]) Shutdown() {
	d.isShutdown.Close()
	d.Wait()
}

// Wait blocks until the debouncer exits
//   - the debouncer exits from input channel closing or Shutdown
func (d *Debouncer[T]) Wait() {
	<-d.in.inputExit.Ch()
	<-d.out.outputExit.Ch()
}

// inputThread debounces the input channel until it closes or Shutdown
func (d *debouncerIn[T]) inputThread() {
	defer d.inputExit.Close()
	defer Recover(func() DA { return A() }, NoErrp, d.errorSink)
	defer d.maxDelayTimer.Stop()
	defer d.debounceTimer.Stop()
	defer d.buffer.EmptyCh() // close of buffer causes output thread to eventually exit

	// debounce timer was started
	var debounceTimerRunning bool

	// read input channel and save values to unbound buffer
	for {

		// wait for value or shutdown
		var value T
		var hasValue bool
		select {
		case value, hasValue = <-d.inputCh:
			if hasValue {
				break // a value was received
			}
			return // the input channel closed return
		case <-d.isShutdown.Ch():
			return // shutdown received return
		}

		// put read value in unbound buffer
		d.buffer.Send(value)

		// a value was received. If max delay is used and not running,
		// start it
		if d.useMaxDelay && d.maxDelayRunning.CompareAndSwap(false, true) {
			d.maxDelayTimer.Reset(0)
		}

		// if debounce timer is used,
		// start or extend debounce timer
		if d.debounceInterval > 0 {
			if debounceTimerRunning {
				// get debounceTimer ready for reset
				d.debounceTimer.Stop()
				select {
				case <-d.debounceTimer.C:
				default:
				}
			} else {
				debounceTimerRunning = true
			}
			// Reset should be invoked only on:
			//	- stopped or expired timers
			//	- with drained channels
			d.debounceTimer.Reset(d.debounceInterval)
		}
	}
}

// outputThread copies the unbound buffer to sender whenever
// a timer expires
func (d *debouncerOut[T]) outputThread() {
	defer d.isShutdown.Close() // shutdown input thread if running
	defer d.outputExit.Close()
	defer Recover(func() DA { return A() }, NoErrp, d.errorSink)

	// while buffer is not closed and emptied, wait for:
	//	- debounce timer expired triggering send,
	//	- maxDelay timer expired triggering send,
	//	- input thread exiting or
	//	- shutdown causing exit
	for !IsClosed[T](d.buffer) {
		select {
		// input thread starts and extends the debounce timer as
		// values are received
		//	- if it expires due to long time between incoming values,
		//		it triggers a send here
		case <-d.debounceC:
		// input thread starts the max delay timer upon receining a value
		// and it is not running
		//	- if it expires prior to debounce timer, it triggers a send here
		case <-d.maxDelayTimer.C: // send due to max Delay reached
		case <-d.isInputExit: // input thread did exit
		case <-d.isShutdown.Ch():
			return // shutdown received
		}

		// sending values, so reset max delay timer
		if d.useMaxDelay && d.maxDelayRunning.Load() {
			d.maxDelayTimer.Stop()
			d.maxDelayRunning.Store(false)
		}

		// send any values
		if values := d.buffer.GetAll(); len(values) > 0 {
			d.sender(values)
		}
	}
}
