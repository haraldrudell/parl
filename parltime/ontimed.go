/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parltime

import (
	"context"
	"time"

	"github.com/haraldrudell/parl"
)

// OnTimed provides events to perform on the hour
type OnTimed struct {
	nowChannel chan time.Time
	sdChan     chan struct{}
	isClose    parl.AtomicBool
	ctx        context.Context
}

func GetTimer(period time.Duration, timeZone Tz, ctx context.Context) (ot *OnTimed) {
	if ctx == nil {
		ctx = context.Background()
	}
	ot = &OnTimed{nowChannel: make(chan time.Time), sdChan: make(chan struct{}), ctx: ctx}
	if period == 0 {
		return
	}
	go ot.run(period, timeZone)
	return
}

// NewOnTimedLocal provides events to perform on the hour
func NewOnTimedLocal(period time.Duration) (ot *OnTimed) {
	return GetTimer(period, LOCAL, nil)
}

// NewOnTimed provides events to perform on the hour
func NewOnTimed(period time.Duration) (ot *OnTimed) {
	return GetTimer(period, UTC, nil)
}

// NewOnTimed initializes OnTimed
func (ot *OnTimed) run(period time.Duration, timeZone Tz) {
	defer parl.Recover("OnTimed.run", nil, func(err error) { parl.Info("%+v\n", err) })
	var timer *time.Timer
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()
	for {
		timer = OnTimer(period, timeZone)
		select {
		case <-ot.ctx.Done():
		case <-ot.sdChan:
		case now := <-timer.C:
			ot.nowChannel <- now
			continue
		}
		break
	}
}

// Chan gets a channel to wait for
func (ot *OnTimed) Chan() (ch <-chan time.Time) {
	return ot.nowChannel
}

// Close stop the durations
func (ot *OnTimed) Close() {
	if ot.isClose.Set() {
		close(ot.sdChan)
	}
}
