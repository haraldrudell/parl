/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"time"

	"github.com/haraldrudell/parl"
)

// OnTimedThread returns an OnTimed object that send time values by calendar period.
// period is the calendar operiod such as time.Hour.
// timeZone is time zone such as time.Local or time.UTC.
func OnTimedThread(send func(time.Time), period time.Duration, loc *time.Location, g0 parl.Go) {
	var err error
	defer g0.Done(&err)
	defer parl.Recover(parl.Annotation(), &err, parl.NoOnError)

	timer := OnTimer(period, time.Now().In(loc))
	defer timer.Stop()

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
			return
		case t := <-C:
			if ticker == nil {
				ticker = time.NewTicker(period)
				C = ticker.C
			}
			send(t)
		}
	}
}
