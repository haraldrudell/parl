/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"time"

	"github.com/haraldrudell/parl/recover"
)

// OnTimedThread invokes a callback on period-multiples since zero-time. thread-safe.
// OnTimedThread is invoked in a go statement, running in its own thread.
// The OnTimedThread thread invokes the send callback function periodically.
//   - the send function callback needs to be thread-safe and is invoked with the trig time provided
//   - loc is scheduling time zone for durations 24 h or greater eg. time.Local or time.UTC
//   - the timer thread is cancelled via g0.Context
//   - OnTimedThread uses g0.Done to provide thread error and to be waited-upon
//   - period must be greater than zero or panic
//   - if the send callback panics, that is a g0 fatal error
//
// Usage:
//
//	gc := g0.NewGoGroup(context.Background())
//	defer gc.Wait()
//	defer gc.Cancel()
//	go ptime.OnTimedThread(someFunc, time.Second, time.Local, gc.Add(parl.EcSharedChan, parl.ExCancelOnExit).Go())
//	…
func OnTimedThread(send func(at time.Time), period time.Duration, loc *time.Location, g0 recover.Go) {
	var err error
	g0.AddError(nil)
	defer g0.Done(&err)
	defer recover.Recover(recover.Annotation(), &err, recover.NoOnError)

	// timer is a time.Timer delaying until the first trig point
	timer := OnTimer(period, time.Now().In(loc))
	defer timer.Stop()

	// ticker is a time.Ticker that provides subsequent trig events
	var ticker *time.Ticker
	defer func() {
		if ticker != nil {
			ticker.Stop()
		}
	}()

	done := g0.Context().Done()
	C := timer.C
	for {
		select {
		case <-done:
			return // g0.Context cancel exit
		case t := <-C: // period trigged with its time.Time value
			if ticker == nil {
				ticker = time.NewTicker(period)
				C = ticker.C
			}
			send(t) // invoke callback
		}
	}
}
