/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"context"
	"time"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
)

// OnTicker is a ticker triggering on period-multiples since zero time
type OnTicker struct {
	C       <-chan time.Time // The channel on which the ticks are delivered.
	onTickp *onTick
}

type onTick struct {
	period               time.Duration
	callback             func(at time.Time)
	durationUntilFirstOn time.Duration
	c                    chan time.Time
	ctx                  context.Context
	cancel               context.CancelFunc
}

// NewOnTicker returns a ticker that ticks on period-multiples since zero time.
//   - loc is optional time zone, default time.Local, other time zone is time.UTC
//   - time zone matters for periods over 1 hour
//   - period must be greater than zero or panic
//   - OnTicker is a time.Ticker using on-period multiples
func NewOnTicker(period time.Duration, loc ...*time.Location) (onTicker *OnTicker) {
	var loc0 *time.Location
	if len(loc) > 0 {
		loc0 = loc[0]
	}
	if loc0 == nil {
		loc0 = time.Local
	}
	return NewOnTicker2(period, loc0, nil, nil)
}

func NewOnTicker2(period time.Duration, loc *time.Location, callback func(at time.Time), g0 Go) (onTicker *OnTicker) {

	// create local onTick struct
	ot := onTick{
		period:               period,
		callback:             callback,
		durationUntilFirstOn: Duro(period, time.Now().In(loc)),
		c:                    make(chan time.Time, 1),
	}
	var ctx context.Context
	if g0 != nil {
		ctx = g0.Context()
	} else {
		ctx = context.Background()
	}
	ot.ctx, ot.cancel = context.WithCancel(ctx)
	go ot.onTickerThread(g0)

	return &OnTicker{C: ot.c, onTickp: &ot}
}

func (o *OnTicker) Stop() {
	o.onTickp.cancel() // signal cancel to thread
}

func (o *onTick) onTickerThread(g0 Go) {
	var err error
	if g0 != nil {
		defer g0.Done(&err)
		defer cyclebreaker.Recover(cyclebreaker.Annotation(), &err, cyclebreaker.NoOnError)
	} else {
		defer cyclebreaker.Recover(cyclebreaker.Annotation(), &err, cyclebreaker.Infallible)
	}
	defer o.cancel()

	var timer = time.NewTimer(o.durationUntilFirstOn)
	defer timer.Stop()

	done := o.ctx.Done()
	C := timer.C
	var t time.Time
	select {
	case <-done:
		return // context canceled return
	case t = <-C:
		o.sendTime(t)
	}

	var ticker = time.NewTicker(o.period)
	defer ticker.Stop()

	C = ticker.C
	for {
		select {
		case <-done:
			return // context canceled return
		case t = <-C:
			o.sendTime(t)
		}
	}
}

func (o *onTick) sendTime(t time.Time) {
	select {
	case o.c <- t:
	default:
	}
	if callback := o.callback; callback != nil {
		callback(t)
	}
}

type Go interface {
	AddError(err error)
	Done(err *error)
	Context() (ctx context.Context)
}
