/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"sync"
	"time"
)

// runner is a container executing rate-counters for a particular interval
type runner struct {
	lock  sync.RWMutex
	tasks []RateRunnerTask // behind lock
}

// runner returns a container for rate-counters of a particular interval
func NewRunner() (run *runner) {
	return &runner{}
}

// Add adds an additional rate-counter to this container
func (r *runner) Add(task RateRunnerTask) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.tasks = append(r.tasks, task)
}

// Do is invoked by a timer with an intended at time
func (r *runner) Do(at time.Time) {
	_ = at
	at = time.Now() // obtain accurate time-stamp
	for i := 0; ; i++ {
		task := r.task(i)
		if task == nil {
			return
		}
		task.Do(at)
	}
}

// task returns task n 0…, nil if no such task exists
func (r *runner) task(i int) (task RateRunnerTask) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if i < 0 || i >= len(r.tasks) {
		return
	}
	task = r.tasks[i]
	return
}
