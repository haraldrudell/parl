/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"sync"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/ptime"
)

type RateRunner struct {
	g0 parl.GoGen

	lock  sync.Mutex
	subGo parl.SubGo
	m     map[time.Duration]*runner
}

func NewRateRunner(g0 parl.GoGen) (rr *RateRunner) {
	return &RateRunner{
		g0: g0,
		m:  map[time.Duration]*runner{},
	}
}

type RateRunnerTask interface {
	Do()
}

func (rr *RateRunner) AddTask(period time.Duration, task RateRunnerTask) {
	rr.lock.Lock()
	defer rr.lock.Unlock()

	if runner, ok := rr.m[period]; ok {
		runner.Add(task)
		return
	}

	if rr.g0 == nil {
		panic(perrors.NewPF("RateCounters instantiated with parl.Go nil"))
	} else if rr.subGo == nil {
		rr.subGo = rr.g0.SubGo()
	}

	runner := NewRunner()
	runner.Add(task)
	go ptime.OnTimedThread(runner.Do, period, time.Local, rr.subGo.Go())
	rr.m[period] = runner
}
