/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"time"
)

const (
	defaultPeriod = time.Second
)

type Periodically struct {
	period time.Duration
	fn     func(t time.Time)
	ctx    context.Context
}

func NewPeriodically(fn func(t time.Time), ctx context.Context, period ...time.Duration) (periodically *Periodically) {
	p := Periodically{ctx: ctx, fn: fn}
	if len(period) > 0 {
		p.period = period[0]
	}
	if p.period < defaultPeriod {
		p.period = defaultPeriod
	}
	go p.doThread()
	return &p
}

func (p *Periodically) doThread() {
	defer Recover(Annotation(), nil, Infallible)

	ticker := time.NewTicker(p.period)
	defer ticker.Stop()

	done := p.ctx.Done()
	for {
		select {
		case <-done:
			return // context cancel exit
		case t := <-ticker.C:
			go p.doFn(t)
		}
	}
}

func (p *Periodically) doFn(t time.Time) {
	defer Recover(Annotation(), nil, Infallible)

	p.fn(t)
}
