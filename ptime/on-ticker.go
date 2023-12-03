/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors"
)

// OnTicker is a ticker triggering on period-multiples since zero time
//   - [time.Ticker] has a C field and Stop and Reset methods
//   - onTickerThread is launched in the new function
//   - — to avoid memory leaks, the thread cannot hold pointers to OnTicker
//   - — but the thread and OnTicker can both point to shared data onTick
type OnTicker struct {
	C       <-chan time.Time // The channel on which the ticks are delivered.
	onTickp *onTick
}

// onTick holds data shared between an OnTicker instance and its thread
type onTick struct {
	period       time.Duration
	loc          *time.Location
	ticker       *time.Ticker
	isThreadExit atomic.Bool
	wg           sync.WaitGroup
	cancel       context.CancelFunc
	ctx          context.Context
}

// NewOnTicker returns a ticker that ticks on period-multiples since zero time.
//   - loc is optional time zone, default time.Local, other time zone is time.UTC
//   - — time zone matters for periods longer than 1 hour or time-zone offiset minutes
//   - period must be greater than zero or panic
//   - OnTicker is a time.Ticker enhanced with on-period multiples
func NewOnTicker(period time.Duration, loc ...*time.Location) (onTicker *OnTicker) {
	var loc0 *time.Location
	if len(loc) > 0 {
		loc0 = loc[0]
	}
	if loc0 == nil {
		loc0 = time.Local
	}
	// period panic in new function, not in thread
	if period <= 0 {
		panic(perrors.NewPF("period must be positive"))
	}

	// create onTick
	var o = onTick{
		period: period,
		loc:    loc0,
		ticker: time.NewTicker(time.Hour),
	}
	// ticker is a stopped ticker with empty channel
	o.ticker.Stop()
	o.ctx, o.cancel = context.WithCancel(context.Background())

	// launch thread
	o.wg.Add(1)
	go o.onTickerThread()

	return &OnTicker{C: o.ticker.C, onTickp: &o}
}

func (o *OnTicker) Stop() {
	o.onTickp.cancel()      // signal to thread to cancel
	o.onTickp.ticker.Stop() // ensure ticker is stopped
	o.onTickp.wg.Wait()     // wait for thread to exit
}

// onTickerThread aligns o.ticker with on-period, then exits
func (o *onTick) onTickerThread() {
	defer o.isThreadExit.Store(true)
	defer o.wg.Done()
	var err error
	defer cyclebreaker.Recover(func() cyclebreaker.DA { return cyclebreaker.A() }, &err, cyclebreaker.Infallible)

	var timer = time.NewTimer(Duro(o.period, time.Now().In(o.loc)))
	defer timer.Stop()

	//	- 1. wait for on-period time
	//	- 2. send a fake initial tick
	//	- 3. start ticker with correct period
	//	- 4. exit thread
	done := o.ctx.Done()
	Cin := timer.C
	// convert read channel o.ticker.C to a send channel
	Cout := *(*chan<- time.Time)(unsafe.Pointer(&o.ticker.C))
	var t time.Time
	select {
	case <-done: // context canceled
	case t = <-Cin:
		o.ticker.Reset(o.period) // start the ticker
		Cout <- t                // send fake initial tick
	}
}
