/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors"
)

// ClosingTicker is like time.Ticker but the channel C closes on shutdown.
// A closing channel is detectable by listening threads.
// If the computer is busy swapping, ticks will be lost.
// MaxDuration indicates the longest time observed between ticks.
// MaxDuration is normally 1 s
type ClosingTicker struct {
	C                 <-chan time.Time
	MaxDuration       time.Duration // atomic int64
	perrors.ParlError               // panic in Shutdown or panic in tick thread
	isShutdownRequest chan struct{}
	isTickThreadExit  chan struct{}
	shutdownOnce      sync.Once
}

// NewClosingTicker returns a Ticker whose channel C closes on shutdown.
// A closing channel is detectable by a listening thread.
func NewClosingTicker(d time.Duration) (t *ClosingTicker) {
	ch := make(chan time.Time)
	ticker := time.NewTicker(d)
	t0 := ClosingTicker{
		C:                 ch,
		isShutdownRequest: make(chan struct{}),
		isTickThreadExit:  make(chan struct{}),
	}

	// luanch our tick thread
	go t0.tick(ch, ticker)

	return &t0
}

// Shutdown causes the channel C to close and resources to be released
func (t *ClosingTicker) Shutdown() {
	t.shutdownOnce.Do(func() {
		defer cyclebreaker.Recover2(func() cyclebreaker.DA { return cyclebreaker.A() }, nil, t.AddErrorProc)

		close(t.isShutdownRequest)
		<-t.isTickThreadExit
	})
}

func (t *ClosingTicker) GetError() (maxDuration time.Duration, err error) {
	maxDuration = time.Duration(atomic.LoadInt64((*int64)(&t.MaxDuration)))
	err = t.ParlError.GetError()
	return
}

func (t *ClosingTicker) tick(out chan time.Time, ticker *time.Ticker) {
	defer close(t.isTickThreadExit)
	defer cyclebreaker.Recover2(func() cyclebreaker.DA { return cyclebreaker.A() }, nil, t.AddErrorProc)
	defer close(out)
	defer ticker.Stop()

	var last time.Time
	var maxDuration time.Duration

	C := ticker.C
	isShutdown := t.isShutdownRequest
	for {
		select {
		case timeValue := <-C:
			out <- timeValue
			if !last.IsZero() {
				d := timeValue.Sub(last)
				if d > maxDuration {
					maxDuration = d
					atomic.StoreInt64((*int64)(&t.MaxDuration), int64(maxDuration))
				}
			}
			last = timeValue
			continue
		case <-isShutdown:
		}
		break
	}
}
