/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"math"
	"time"

	"github.com/haraldrudell/parl"
)

// runner is a container executing rate-counters for a particular interval
type runner struct {
	// lock makes tasks thread-safe
	lock parl.RWMutex
	// behind lock
	tasks []RateRunnerTask
}

// runner returns a container for rate-counters of a particular interval
func NewRunner() (run *runner) { return &runner{} }

// Add adds an additional rate-counter to this container
func (r *runner) Add(task RateRunnerTask) {
	defer r.lock.Lock().Unlock()

	r.tasks = append(r.tasks, task)
}

// Do is invoked by a timer with an intended at time
func (r *runner) Do(at time.Time) {
	_ = at

	// obtain accurate time-stamp
	at = time.Now()

	for i := range math.MaxInt {
		var task = r.task(i)
		if task == nil {
			return
		}
		task.Do(at)
	}
}

// task returns task n 0…, nil if no such task exists
func (r *runner) task(i int) (task RateRunnerTask) {
	defer r.lock.RLock().RUnlock()

	// bad task ID
	if i < 0 || i >= len(r.tasks) {
		return // bad task ID return
	}

	task = r.tasks[i]
	return
}
