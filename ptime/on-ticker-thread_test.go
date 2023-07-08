/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/haraldrudell/parl/perrors"
)

type OnTickerThreadTester struct {
	ats   []time.Time
	dones []error
	wg    sync.WaitGroup
	done  sync.WaitGroup
	ctx   context.Context
}

func NewOnTickerThreadTester(ctx context.Context, ticks int) (o *OnTickerThreadTester) {
	ox := OnTickerThreadTester{ctx: ctx}
	ox.wg.Add(ticks)
	ox.done.Add(1)
	return &ox
}
func (o *OnTickerThreadTester) callback(at time.Time) {
	o.ats = append(o.ats, at)
	o.wg.Done()
}
func (o *OnTickerThreadTester) Done(errp *error) {
	defer o.done.Done()

	var err error
	if errp != nil {
		err = *errp
	}
	o.dones = append(o.dones, err)
}
func (o *OnTickerThreadTester) Context() (ctx context.Context) {
	return o.ctx
}

func TestOnTickerThread(t *testing.T) {
	var period = time.Millisecond
	var ticks = 2

	var ctx, cancel = context.WithCancel(context.Background())
	var o = NewOnTickerThreadTester(ctx, ticks)
	go OnTickerThread(o.callback, period, nil, o)

	// wait for 2 ticks
	o.wg.Wait()

	// cancel the ticker
	cancel()
	// wait for thread to exit
	o.done.Wait()

	// thread should have exited without error
	if e := o.dones[0]; e != nil {
		t.Errorf("thread err: %s", perrors.Short(e))
	}
	// there should be 2 ticks
	if len(o.ats) != ticks {
		t.Fatalf("ticks: %d exp %d", len(o.ats), ticks)
	}
	// duration should relate to period
	var duration = o.ats[1].Sub(o.ats[0])
	if duration < period/2 {
		t.Errorf("duration too short: %s exp %s", duration, period)
	} else if duration >= 2*period {
		t.Errorf("duration too long: %s exp %s", duration, period)
	}
}
