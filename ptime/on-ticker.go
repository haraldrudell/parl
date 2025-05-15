/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"sync"
	"time"
	"unsafe"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors"
)

// OnTicker is a ticker triggering on period-multiples since zero time
//   - [time.Ticker] has a C field and Stop method
//   - onTickerThread is launched in the new function
//   - as of go1.23, trig is sent on unbuffered channel
//   - to combine a timer then ticker, OnTicker needs a thread
//   - once on-period is attained, the thread exits
//   - the consuming thread reads channel OnTicker.C delegated to [time.Timer.C]
type OnTicker struct {
	// C is the channel on which the ticks are delivered
	//	- must be a field named C
	//	- delegated to ticker channel
	C <-chan time.Time
	// period is the set period, 1 ns or greater
	period time.Duration
	// timeLocation handles time zone issues
	//	- typically pointer to static variable
	timeLocation *time.Location
	// ticker runs subsequent ticks after on-period attained
	//	- ticker should not be copied: pointer to heap
	//	- [time.NewTicker] costs channel and heap-struct allocation
	ticker *time.Ticker
	// orderThreadToExit exits the thread
	//	- must be closing channel
	//	- Stop can be Once + closing channel
	orderThreadToExit cyclebreaker.Awaitable
	// wg makes thread awaitable
	wg sync.WaitGroup
}

// NewOnTicker returns a ticker that ticks on period-multiples since zero time.
//   - period: when channel emits ticks, on-period, on-the-hour etc.
//   - — must be greater than zero or panic
//   - loc: optional time zone, default time.Local, other time zone is time.UTC
//   - — time zone matters for periods longer than 1 hour or for location with time-zone offset minutes
//   - —
//   - OnTicker is a time.Ticker enhanced to trig at on-period multiples
//   - [OnTicker.C] unbuffered channel delegated to ticker
//   - [OnTicker.Stop] releases resources, must be invoked for every NewOnTicker
func NewOnTicker(period time.Duration, loc ...*time.Location) (onTicker *OnTicker) {
	return NewOnTickerp(nil, period, loc...)
}

// NewOnTickerp returns a ticker that ticks on period-multiples since zero time.
//   - period: when channel emits ticks, on-period, on-the-hopur etc.
//   - — must be greater than zero or panic
//   - loc: optional time zone, default time.Local, other time zone is time.UTC
//   - — time zone matters for periods longer than 1 hour or for location with time-zone offset minutes
//   - —
//   - OnTicker is a time.Ticker enhanced to trig at on-period multiples
//   - [OnTicker.C] unbuffered channel delegated to ticker
//   - [OnTicker.Stop] releases resources, must be invoked for every NewOnTicker
func NewOnTickerp(fieldp *OnTicker, period time.Duration, loc ...*time.Location) (onTicker *OnTicker) {

	// get onTicker
	if fieldp != nil {
		onTicker = fieldp
	} else {
		onTicker = &OnTicker{}
	}

	// period panic in new function, not in thread
	if period <= 0 {
		panic(perrors.NewPF("period must be positive"))
	}
	// ticker is a stopped ticker with empty channel
	var ticker = time.NewTicker(time.Hour)
	ticker.Stop()

	*onTicker = OnTicker{
		C:      ticker.C,
		period: period,
		ticker: ticker,
	}
	if len(loc) > 0 {
		onTicker.timeLocation = loc[0]
	}
	if onTicker.timeLocation == nil {
		onTicker.timeLocation = time.Local
	}

	// launch thread
	onTicker.wg.Add(1)
	go onTicker.onTickerThread()
	return
}

// Stop releases resources
func (o *OnTicker) Stop() {

	// uninitialized would panic with nil o.ticker
	if o.period == 0 {
		panic(perrors.NewPF("uninitialized OnTicker"))
	}

	// signal to thread to cancel: idempotent
	o.orderThreadToExit.Close()
	// ensure ticker is stopped: idempotent
	//	- thread may have already reset ticker and exited
	o.ticker.Stop()
	// wait for thread to exit: idempotent
	o.wg.Wait()
}

// onTickerThread aligns o.ticker with on-period, then exits
//   - Stop invocation may exit thread prematurely
//   - errors are logged to standard error: should not be any
func (o *OnTicker) onTickerThread() {
	defer o.wg.Done()
	if o.orderThreadToExit.IsClosed() {
		return
	}
	var err error
	defer cyclebreaker.Recover(func() cyclebreaker.DA { return cyclebreaker.A() }, &err, cyclebreaker.Infallible)

	// timer runs until first on-period
	var timer = time.NewTimer(Duro(o.period, time.Now().In(o.timeLocation)))
	defer timer.Stop()

	//	- 1. wait for on-period time
	//	- 2. send a fake initial tick
	//	- 3. start ticker with correct period
	//	- 4. exit thread
	var a = o.orderThreadToExit.Ch()
	var t time.Time
	select {
	case <-a:
		// Stop invocation
		return
	case t = <-timer.C:
	}
	// now on-period

	// start the ticker
	o.ticker.Reset(o.period)
	// convert read channel o.ticker.C to a send channel
	var Cout = *(*chan<- time.Time)(unsafe.Pointer(&o.ticker.C))

	// send fake initial tick
	select {
	// may block here
	case Cout <- t: // sent fake tick
	case <-a: // cancel from Stop invocation
	}
}
