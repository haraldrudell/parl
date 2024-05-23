/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"context"
	"time"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors"
)

type Go interface {
	Done(err *error)
	Context() (ctx context.Context)
}

// OnTickerThread is a goroutine that invokes a callback periodically
//   - period must be greater than zero or panic
//   - loc is optional time zone, default time.Local, other time zone is time.UTC
//   - — time zone matters for periods longer than 1 hour or time-zone offiset minutes
//   - g can be nil, in which panics are echoed to standard error.
//     The thread is then not awaitable or cancelable
//   - OnTickerThread is a time.Ticker using on-period multiples that invokes
//     a callback periodically without further action
//   - OnTickerThread eliminates channel receive actions at
//     the cost of an additional OnTickerThread thread
//   - a thread-less on-period ticker is OnTicker
func OnTickerThread(callback func(at time.Time), period time.Duration, loc *time.Location, g Go) {
	var err error
	if g != nil {
		defer g.Done(&err)
		defer cyclebreaker.Recover(func() cyclebreaker.DA { return cyclebreaker.A() }, &err)
	} else {
		defer cyclebreaker.Recover(func() cyclebreaker.DA { return cyclebreaker.A() }, &err, cyclebreaker.Infallible)
	}

	if callback == nil {
		err = perrors.NewPF("callback cannot be nil")
		return
	}

	var ticker = NewOnTicker(period, loc)
	defer ticker.Stop()

	C := ticker.C
	var done <-chan struct{}
	if g != nil {
		done = g.Context().Done()
	}
	var t time.Time
	for {
		select {
		case <-done:
			return // context canceled return
		case t = <-C:
		}

		callback(t)
	}
}
