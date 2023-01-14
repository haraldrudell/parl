/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"time"

	"github.com/haraldrudell/parl"
)

// OnTimedThread invokes a callback on period-multiples since zero-time.
//   - send is a thread-safe callback invoked on the schedule with the trig time provided
//   - loc contains time zone for durations 24 h or greater eg. time.Local or time.UTC
//   - the timer is cancelled using g0.Context
//   - OnTimedThread uses g0.Done to provide thread error and to be waited-upon
//   - period must be greater than zero or panic
//
// Usage:
//
//	gc := g0.NewGoGroup(context.Background())
//	defer gc.Wait()
//	defer gc.Cancel()
//	go ptime.OnTimedThread(someFunc, time.Second, time.Local, gc.Add(parl.EcSharedChan, parl.ExCancelOnExit).Go())
//	…
func OnTimedThread(send func(at time.Time), period time.Duration, loc *time.Location, g0 parl.Go) {
	var err error
	defer g0.Register().Done(&err)
	defer parl.Recover(parl.Annotation(), &err, parl.NoOnError)

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
