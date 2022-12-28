/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"time"
)

type runner struct {
	tasks []RateRunnerTask
}

func NewRunner() (run *runner) {
	return &runner{}
}

func (r *runner) Add(task RateRunnerTask) {
	r.tasks = append(r.tasks, task)
}

func (r *runner) Do(at time.Time) {
	for _, task := range r.tasks {
		task.Do()
	}
}
