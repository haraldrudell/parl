/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/ptime"
)

// RateRunner is a container managing threads executing rate-counter tasks by their period
type RateRunner struct {
	// goGen creates rate-runner threads for rate-counters
	goGen parl.GoGen
	// lock makes m thread-safe
	lock parl.Mutex
	// written behind lock
	subGo parl.SubGo
	// m contains runner threads
	//	- key: a rate-countger duration
	//	- value: runner struct associated with a running thread
	// behind lock
	m map[time.Duration]*runner
}

// NewRateRunner returns a thread-container for running rate-counter averaging
//   - g: goGen creates rate-runner threads for any rate-counters
func NewRateRunner(goGen parl.GoGen, fieldp ...*RateRunner) (r *RateRunner) {

	// get r
	if len(fieldp) > 0 {
		r = fieldp[0]
	}
	if r == nil {
		r = &RateRunner{}
	}

	*r = RateRunner{
		goGen: goGen,
		m:     map[time.Duration]*runner{},
	}
	return
}

// AddTask adds a new rate-counter to the container
//   - period: the required period for particular rate counter
//   - —
//   - for rate counters to cature increasomg or decreasing rate,
//     a value must be inspected periodically
//   - each task thread carries out that inspection on a particular
//     schedule
//   - task is added to slice so must be on heap
func (r *RateRunner) AddTask(period time.Duration, task RateRunnerTask) {
	defer r.lock.Lock().Unlock()

	// check for existing runner with the sought period
	if runner, ok := r.m[period]; ok {
		runner.Add(task)
		return // added to existing runner return
	}

	// ensure subGo
	if r.goGen == nil {
		panic(perrors.NewPF("RateCounters instantiated with parl.Go nil"))
	} else if r.subGo == nil {
		r.subGo = r.goGen.SubGo()
	}

	// start a new runner
	//	- runner is added to map so must be on heap
	var runner = NewRunner()
	runner.Add(task)
	go ptime.OnTickerThread(runner.Do, period, time.Local, r.subGo.Go())
	r.m[period] = runner
}
