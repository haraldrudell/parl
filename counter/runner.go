/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"sync"
	"time"
)

type runner struct {
	lock  sync.RWMutex
	tasks []RateRunnerTask
}

func NewRunner() (run *runner) {
	return &runner{}
}

func (r *runner) Add(task RateRunnerTask) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.tasks = append(r.tasks, task)
}

func (r *runner) Do(at time.Time) {
	for i := 0; ; i++ {
		task := r.task(i)
		if task == nil {
			return
		}
		task.Do()
	}
}

func (r *runner) task(i int) (task RateRunnerTask) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if i < 0 || i >= len(r.tasks) {
		return
	}
	task = r.tasks[i]
	return
}
